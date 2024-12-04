package db

import (
	"ocrserver/config"

	"go.mongodb.org/mongo-driver/mongo"
)

func SelectTable(client *mongo.Client) *mongo.Collection {
	// 데이터베이스 및 컬렉션 선택
	database := client.Database(config.AppConfig.MongoDB.Database)
	collection := database.Collection("User")

	return collection
}
