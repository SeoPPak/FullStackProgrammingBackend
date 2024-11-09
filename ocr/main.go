package main

import (
	"fmt"
	api "ocr/api"
	image "ocr/images"
)

func main() {
	data, contentTpye := image.GetImage()
	//fmt.Printf("from ocr/main: %s\n\n", data)
	res := api.RequsetOCR(data, contentTpye)
	fmt.Printf(string(res))
}
