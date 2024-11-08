package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

func CheckConnection(client *mongo.Client) {
	// 연결 확인
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("MongoDB 연결 실패:", err)
	} else {
		fmt.Println("MongoDB에 성공적으로 연결되었습니다!")
	}
}
