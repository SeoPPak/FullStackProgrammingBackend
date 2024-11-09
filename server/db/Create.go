package db

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

type ExampleDocument struct {
	uid      string
	nickname string
	pw       string
	Record   struct {
		rid    int
		rname  string
		martID int
	}
}

func InsertDocument(collection *mongo.Collection) {
	user := ExampleDocument{
		uid:      "00",
		nickname: "honggil-dong",
		pw:       "qwer1234",
	}
	insertResult, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted document ID:", insertResult.InsertedID)
}
