package image

import (
	"encoding/base64"
	"io/ioutil"
	"log"
)

func GetImage() string {
	path := "images/receipt.jpg"
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return ""
	}
	var base64Encoding string
	/*
		mimeType := http.DetectContentType(bytes)
		switch mimeType {
		case "image/jpeg":
			base64Encoding += "data:image/jpeg;base64,"
		case "image/png":
			base64Encoding += "data:image/png;base64,"
		}
	*/
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)
	return base64Encoding
}
