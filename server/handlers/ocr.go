package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/config"
	"server/images"
	"server/models"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

func RequestOCR(c *gin.Context) {
	ocrURl := config.AppConfig.OCR.URL
	ocrSecretKey := config.AppConfig.OCR.SecretKey

	images.GetImage(c)

	timestamp := int(time.Now().Unix())

	session := c.MustGet("session").(*sessions.Session)
	ocrImages := session.Values["ocrImages"].([]models.OCRImage)

	log.Printf("OCR Request: %s\n", ocrImages)

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
	res, _ := client.Do(req)

	var data map[string]interface{}
	merr := json.NewDecoder(res.Body).Decode(&data)
	if merr != nil {
		log.Fatal(merr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode OCR response"})
	}

	defer res.Body.Close()

	log.Printf("OCR Result: %s\n", data)
	c.JSON(http.StatusOK, data)

}
