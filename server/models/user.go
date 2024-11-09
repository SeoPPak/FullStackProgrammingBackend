package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type DBRequest struct {
	Uid      primitive.ObjectID `bson:"uid", omitempty`
	Nickname string             `bson:"nickname"`
	Email    string             `bson:"email"`
	Pw       string             `bson:"pw"`
	Record   DBRecord           `bson:"record"`
	Mart     DBMart             `bson:"mart"`
	Product  []DBProduct        `bson:"product"`
}

type DBRecord struct {
	Rid       int    `bson:"rid"`
	Rname     string `bson:"rname"`
	TimeStamp string `bson:"timeStamp"`
}

type DBProduct struct {
	Pname  string `bson:"pname"`
	Price  int    `bson:"price"`
	Amount int    `bson:"amount"`
}

type DBMart struct {
	Mid         int    `bson:"mid"`
	MartAddress string `bson:"martAddress"`
	MartName    string `bson:"martName"`
}

type LoginRequest struct {
	Email    string `bson:"email"`
	Password string `bson:"password"`
	Nickname string `bson:"nickname"`
}

type User struct {
	Uid      primitive.ObjectID `bson:"uid"`
	Nickname string             `bson:"nickname"`
	Email    string             `bson:"email"`
	Pw       string             `bson:"pw"`
}

type SignupRequest struct {
	Uid      primitive.ObjectID `bson:"uid"`
	Email    string             `bson:"email"`
	Pw       string             `bson:"password"`
	Nickname string             `bson:"name"`
}
