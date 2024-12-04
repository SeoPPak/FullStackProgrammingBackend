package models

type RecordInput struct {
	Uid        string      `bson:"uid"`
	Record     DBRecord    `bson:"record"`
	Mart       DBMart      `bson:"mart"`
	Product    []DBProduct `bson:"product"`
	TotalPrice int         `bson:"totalPrice"`
}

type DBRecord struct {
	Rid       string `bson:"rid"`
	Rname     string `bson:"rname"`
	TimeStamp string `bson:"timeStamp"`
}

type DBProduct struct {
	Pname  string `bson:"pname"`
	Price  int    `bson:"price"`
	Amount int    `bson:"amount"`
}

type DBMart struct {
	MartAddress string `bson:"martAddress"`
	MartName    string `bson:"martName"`
	Tel         string `bson:"tel"`
}

type RecordList struct {
	Uid        string     `bson:"uid"`
	Record     DBRecord   `bson:"record"`
	Mart       SimpleMart `bson:"mart"`
	TotalPrice int        `bson:"totalPrice"`
}

type SimpleMart struct {
	MartName string `bson:"martName"`
}

type UpdateRecordNameRequest struct {
	Rid     string `json:"rid" binding:"required"`
	NewName string `json:"newName" binding:"required"`
}

// UpdateProductPriceRequest represents the request body for updating product price
type UpdateProductPriceRequest struct {
	Rid         string `json:"rid" binding:"required"`
	ProductName string `json:"pName" binding:"required"`
	NewPrice    int    `json:"newPrice" binding:"required"`
}

type GetProductInfoRequest struct {
	Rid   string `json:"rid" binding:"required"`
	Pname string `json:"pname" binding:"required"`
}

type UpdateProductRequest struct {
	Rid       string `json:"rid" binding:"required"`
	Pname     string `json:"pname" binding:"required"`
	NewPrice  int    `json:"price" binding:"required"`
	NewAmount int    `json:"amount" binding:"required"`
}

type UpdateMartRequest struct {
	Rid         string `json:"rid" binding:"required"`
	NewMartName string `json:"newMartName" binding:"required"`
	NewMartAddr string `json:"newMartAddr" binding:"required"`
	NewMartTel  string `json:"newMartTel" binding:"required"`
}

type UpdateRecordRequest struct {
	Rid      string `json:"rid" binding:"required"`
	NewRname string `json:"newRname" binding:"required"`
	NewTime  string `json:"newTime" binding:"required"`
}
