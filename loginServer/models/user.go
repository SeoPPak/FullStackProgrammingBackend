package models

type DBRequest struct {
	Uid      string      `bson:"uid", omitempty`
	Nickname string      `bson:"nickname"`
	Email    string      `bson:"email"`
	Pw       string      `bson:"pw"`
	Record   DBRecord    `bson:"record"`
	Mart     DBMart      `bson:"mart"`
	Product  []DBProduct `bson:"product"`
}

type LoginRequest struct {
	Email    string `bson:"email"`
	Password string `bson:"pw"`
	Nickname string `bson:"nickname"`
}

type User struct {
	Uid      string `bson:"uid"`
	Nickname string `bson:"nickname"`
	Email    string `bson:"email"`
	Pw       string `bson:"pw"`
}

type SignupRequest struct {
	Uid      string `bson:"uid"`
	Email    string `bson:"email"`
	Pw       string `bson:"pw"`
	Nickname string `bson:"name"`
}
