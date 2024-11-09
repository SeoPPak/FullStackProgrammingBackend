package images

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"server/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func GetImage(c *gin.Context) (string, string) {
	path := "images/receipt.jpg"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return "", ""
	}
	var base64Encoding string

	contentType := "None"

	mimeType := http.DetectContentType(bytes)
	switch mimeType {
	case "image/jpeg":
		contentType = "jpg"
	case "image/png":
		contentType = "png"
	}

	base64Encoding += base64.StdEncoding.EncodeToString(bytes)

	ocrImages := make([]models.OCRImage, 1)
	ocrImages[0] = models.OCRImage{
		Format: contentType,
		Data:   base64Encoding,
		Name:   "receipt",
	}

	session := c.MustGet("session").(*sessions.Session)
	session.Values["ocrImages"] = ocrImages

	return base64Encoding, contentType
}
