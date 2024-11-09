package db

import (
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Client     *mongo.Client
	Collection *mongo.Collection
)

func DBInit() {
	Client, _, _ = ConnectDB()
	Collection = SelectTable(Client)
}
