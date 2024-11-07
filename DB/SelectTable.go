package main

import (
	"go.mongodb.org/mongo-driver/mongo"
)

func SelectTable(client *mongo.Client) *mongo.Collection {
	// 데이터베이스 및 컬렉션 선택
	database := client.Database("local")
	collection := database.Collection("User")

	return collection
}
