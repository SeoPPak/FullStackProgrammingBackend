package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	jwt "ocrserver/auth"
	"ocrserver/config"
	"ocrserver/db"
	"ocrserver/images"
	"ocrserver/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OCRResult struct {
	Data  map[string]interface{}
	Error error
}

func processOCRRequest(ocrURL string, ocrSecretKey string, doc []byte) (*OCRResult, error) {
	req, _ := http.NewRequest("POST", ocrURL, strings.NewReader(string(doc)))
	req.Header.Add("Content-type", "application/json")
	req.Header.Add("X-OCR-SECRET", ocrSecretKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &OCRResult{Data: data}, nil
}

func saveToDatabase(ctx context.Context, recordRequest models.RecordInput) error {
	_, err := db.Collection.InsertOne(ctx, recordRequest)
	return err
}

func RequestOCR(c *gin.Context) {
	parameter := make(map[string][]string)
	if err := c.ShouldBindJSON(&parameter); err != nil {
		log.Printf("Error: invalid request format: %s\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 병렬 처리를 위한 WaitGroup 생성
	var wg sync.WaitGroup
	resultChan := make(chan *OCRResult, len(parameter["image"]))
	errorChan := make(chan error, len(parameter["image"]))

	ocrURL := config.AppConfig.OCR.URL
	ocrSecretKey := config.AppConfig.OCR.SecretKey

	// 각 이미지에 대해 고루틴으로 OCR 처리
	for _, imageData := range parameter["image"] {
		wg.Add(1)
		go func(data string) {
			defer wg.Done()

			ocrImages := images.GetImage(c, []string{data})
			timestamp := int(time.Now().Unix())

			ocrRequest := models.OCRRequest{
				Version:   "V2",
				RequestId: uuid.New().String(),
				Timestamp: strconv.Itoa(timestamp),
				Lang:      "ko",
				Images:    ocrImages,
			}

			doc, _ := json.Marshal(ocrRequest)
			result, err := processOCRRequest(ocrURL, ocrSecretKey, doc)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- result
		}(imageData)
	}

	// 고루틴 완료 대기를 위한 고루틴
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// 결과 처리
	var results []*OCRResult
	for result := range resultChan {
		results = append(results, result)
	}

	// 에러 확인
	select {
	case err := <-errorChan:
		if err != nil {
			log.Printf("Error in OCR processing: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	default:
	}

	// 계정 정보 확인
	account, err := jwt.GetAccount(c)
	if err != nil {
		log.Printf("Error: failed to get account: %s\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get account"})
		return
	}

	// 결과 처리 및 데이터베이스 저장
	var dbRequests []models.RecordInput
	for _, result := range results {
		dbRequest := parseOCRResult(result.Data, account.Uid)
		dbRequests = append(dbRequests, dbRequest)
	}

	// 데이터베이스 저장을 위한 고루틴
	errChan := make(chan error, len(dbRequests))
	for _, request := range dbRequests {
		wg.Add(1)
		go func(req models.RecordInput) {
			defer wg.Done()
			if err := saveToDatabase(ctx, req); err != nil {
				errChan <- err
			}
		}(request)
	}

	// 저장 완료 대기
	wg.Wait()
	close(errChan)

	// 저장 중 발생한 에러 확인
	for err := range errChan {
		if err != nil {
			log.Printf("Error saving to database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save data"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "data successfully processed and saved"})
}

func parseOCRResult(data map[string]interface{}, uid string) models.RecordInput {
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
	product := make(chan interface{}, len(products))

	dbProducts := make([]models.DBProduct, len(products))
	dbProductChan := make(chan models.DBProduct, len(products))

	go func(productChan chan interface{}) {
		defer close(productChan)

		p := <-productChan
		product := p.(map[string]interface{})

		productName := product["name"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)
		price := product["price"].(map[string]interface{})["price"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)
		amount := product["count"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)

		intamount, _ := strconv.Atoi(amount)
		intprice, _ := strconv.Atoi(price)

		dbProductChan <- models.DBProduct{
			Pname:  productName,
			Price:  intprice,
			Amount: intamount,
		}
	}(product)

	for _, p := range products {
		product <- p
		productInput := <-dbProductChan
		dbProducts = append(dbProducts, productInput)
	}
	close(dbProductChan)

	totalPrice := result["totalPrice"].(map[string]interface{})["price"].(map[string]interface{})["formatted"].(map[string]interface{})["value"].(string)
	intTotalPrice, _ := strconv.Atoi(totalPrice)

	recordName := recordTime + martName

	record := models.DBRecord{
		Rid:       uuid.NewString(),
		Rname:     recordName,
		TimeStamp: recordTime,
	}
	return models.RecordInput{
		Uid:        uid,
		Record:     record,
		Mart:       martInput,
		Product:    dbProducts,
		TotalPrice: intTotalPrice,
	}
}
