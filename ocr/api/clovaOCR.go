package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type OCRRequest struct {
	Version   string     `json:"version"`
	RequestId string     `json:"requestId"`
	Timestamp string     `json:"timestamp"`
	Lang      string     `json:"lang, omitempty"`
	Images    []OCRImage `json:"images"`
}

type OCRImage struct {
	Format     string   `json:"format"`
	Data       string   `json:"data,omitempty"`
	Url        string   `json:"url,omitempty"`
	Name       string   `json:"name"`
	TemplateId []string `json:"templateId, omitempty"`
}

func RequsetOCR(data, contentType string) []byte {
	ocrURl := "https://3lw4f4mamp.apigw.ntruss.com/custom/v1/35733/81998f2d759c60f8772617c8d9589f4f1f3e83a4f7fca03370e66ebc35487f3a/document/receipt"
	ocrSecretKey := "d2xXSGVlTElyaGV6VGVEcUFIeXh6d09DTWpOVUdaS0s="

	timestamp := int(time.Now().Unix())
	ocrImages := make([]OCRImage, 1)
	ocrImages[0] = OCRImage{
		Format: contentType,
		Data:   data,
		Name:   "receipt",
	}
	ocrRequest := OCRRequest{
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
	resBody, _ := ioutil.ReadAll(res.Body)

	return resBody
}
