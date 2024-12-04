package images

import (
	"ocrserver/models"

	"github.com/gin-gonic/gin"
)

func GetImage(c *gin.Context, base64Encoding []string) []models.OCRImage {
	/*
		bytes, err := ioutil.ReadFile(path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	*/
	//var base64Encoding string

	contentType := "png"

	//base64Encoding += base64.StdEncoding.EncodeToString(bitmap)

	ocrImages := []models.OCRImage{}
	for i := 0; i < len(base64Encoding); i++ {
		ocrImages = append(ocrImages, models.OCRImage{
			Format: contentType,
			Data:   base64Encoding[i],
			Name:   "receipt",
		})
	}

	return ocrImages
}
