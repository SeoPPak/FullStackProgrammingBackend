package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	jwt "server/auth"
	"server/config"
	"server/db"
	"server/images"
	"server/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

func RequestOCR(c *gin.Context) {

	ocrURl := config.AppConfig.OCR.URL
	ocrSecretKey := config.AppConfig.OCR.SecretKey

	parameter := make(map[string][]string)
	if err := c.ShouldBindJSON(&parameter); err != nil {
		log.Printf("Error: invalid request format:  %s\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	/*
		path := "images/receipt.jpg"
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	*/

	bytes := parameter["image"]
	//log.Printf("OCR Request: %s\n", bytes)

	ocrImages := images.GetImage(c, bytes)

	timestamp := int(time.Now().Unix())

	//log.Printf("OCR Request: %s\n", ocrImages)

	ocrRequest := models.OCRRequest{
		Version:   "V2",
		RequestId: uuid.New().String(),
		Timestamp: strconv.Itoa(timestamp),
		Lang:      "ko",
		Images:    ocrImages,
	}

	doc, _ := json.Marshal(ocrRequest)

	req, _ := http.NewRequest("POST", ocrURl, strings.NewReader(string(doc)))
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("X-OCR-SECRET", ocrSecretKey)

	client := &http.Client{}
	res, rerr := client.Do(req)
	if rerr != nil {
		log.Printf("Error: failed to call api: %s\n", rerr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": rerr.Error()})
		return
	}

	var data map[string]interface{}
	merr := json.NewDecoder(res.Body).Decode(&data)
	if merr != nil {
		log.Printf("Error: failed to decode: %s\n", merr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode OCR response"})
	}

	defer res.Body.Close()

	resp := data["images"].([]interface{})[0].(map[string]interface{})
	images := resp["receipt"].(map[string]interface{})
	result := images["result"].(map[string]interface{})

	paymentInfo := result["paymentInfo"].(map[string]interface{})
	paymentDate := paymentInfo["date"].(map[string]interface{})["formatted"].(map[string]interface{})
	pMonth := paymentDate["month"].(string)
	pDay := paymentDate["day"].(string)
	pYear := paymentDate["year"].(string)

	recordTime := pYear + "-" + pMonth + "-" + pDay

	martInfo := result["storeInfo"].(map[string]interface{})

	martName := martInfo["name"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)
	martAddress := martInfo["addresses"].([]interface{})[0].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)
	tel := martInfo["tel"].([]interface{})[0].(map[string]interface{})["text"].(string)

	martInput := models.DBMart{
		MartName:    martName,
		MartAddress: martAddress,
		Tel:         tel,
	}

	subres := result["subResults"].([]interface{})[0].(map[string]interface{})
	products := subres["items"].([]interface{})

	dbProducts := make([]models.DBProduct, len(products))

	i := 0
	for _, p := range products {
		product := p.(map[string]interface{})

		productName := product["name"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)
		price := product["price"].(map[string]interface{})["price"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)
		amount := product["count"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)

		intamount, _ := strconv.Atoi(amount)
		intprice, _ := strconv.Atoi(price)

		dbProducts[i] = models.DBProduct{
			Pname:  productName,
			Price:  intprice,
			Amount: intamount,
		}
		i++
	}

	totalPrice := result["totalPrice"].(map[string]interface{})["price"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)
	intTotalPrice, _ := strconv.Atoi(totalPrice)

	recordName := recordTime + martName

	record := models.DBRecord{
		Rid:       uuid.NewString(),
		Rname:     recordName,
		TimeStamp: recordTime,
	}

	account, err := jwt.GetAccount(c)
	if err != nil {
		log.Printf("Error: failed to get account: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get account"})
		return
	}
	uid := account.Uid

	recordRequest := models.RecordInput{
		Uid:        uid,
		Record:     record,
		Mart:       martInput,
		Product:    dbProducts,
		TotalPrice: intTotalPrice,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	existingRecord := models.DBRequest{}

	err = db.Collection.FindOne(ctx, bson.M{"uid": uid, "record.rid": recordName}).Decode(&existingRecord)
	if err == nil {
		log.Printf("Error: record already registered: %s\n", err)
		c.JSON(http.StatusConflict, gin.H{"error": "Record already registered"})
		return
	}

	_, err = db.Collection.InsertOne(ctx, recordRequest)
	if err != nil {
		log.Printf("Error: failed to insert record: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "data successfully inserted"})

}
