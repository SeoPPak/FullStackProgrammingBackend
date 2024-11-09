package main

import (
	"fmt"
	api "ocr/api"
	image "ocr/images"
)

func main() {
	data := image.GetImage()
	//fmt.Printf("from ocr/main: %s\n\n", data)
	res := api.RequsetOCR(data)
	fmt.Printf(string(res))
}
