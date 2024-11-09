package image

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
)

func GetImage() (string, string) {
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
	return base64Encoding, contentType
}
