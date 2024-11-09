package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type DBRequest struct {
	uid      string      `json:"uid"`
	nickname string      `json:"nickname"`
	email    string      `json:"email"`
	pw       string      `json:"pw"`
	Record   DBRecord    `json:"record"`
	Mart     DBMart      `json:"mart"`
	Product  []DBProduct `json:"product"`
}

type DBRecord struct {
	rid       int    `json:"rid"`
	rname     string `json:"rname"`
	timeStamp string `json:"timeStamp"`
}

type DBProduct struct {
	pname  string `json:"pname"`
	price  int    `json:"price"`
	amount int    `json:"amount"`
}

type DBMart struct {
	mid         int    `json:"mid"`
	martAddress string `json:"martAddress"`
	martName    string `json:"martName"`
}

func InsertUser(c *gin.Context, collection *mongo.Collection) {

	mart := DBMart{
		mid:         0,
		martAddress: "서울시 강남구",
		martName:    "이마트",
	}

	product := make([]DBProduct, 1)
	product[0] = DBProduct{
		pname:  "사과",
		price:  1000,
		amount: 10,
	}

	record := DBRecord{
		rid:       0,
		rname:     "2021-01-01",
		timeStamp: "2021-01-01 00:00:00",
	}

	user := DBRequest{
		uid:      "00",
		nickname: "honggil-dong",
		email:    "qwer1234@example.com",
		pw:       "qwer1234",
		Record:   record,
		Mart:     mart,
		Product:  product,
	}

	insertResult, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted document ID:", insertResult.InsertedID)
}
