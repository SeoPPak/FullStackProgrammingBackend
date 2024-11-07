package main

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() (*mongo.Client, context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// MongoDB 클라이언트 옵션 설정
	clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017/") // 로컬 MongoDB URI

	// MongoDB 클라이언트 생성 및 연결
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// 연결 확인
	CheckConnection(client)

	return client, ctx, cancel
}

func DisconnectDB(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		log.Fatal(err)
	}
}
