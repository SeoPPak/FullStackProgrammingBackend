package handlers

import (
	"context"
	"log"
	"net/http"
	jwt "server/auth"
	"server/db"
	"server/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func SearchByUid(c *gin.Context) {
	token, err := jwt.GetToken(c)
	if err != nil {
		log.Printf("Token error: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	claims, err := jwt.ValidateToken(token)
	if err != nil {
		log.Printf("Claims error: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	account := claims.Account
	c, err = jwt.SetAccount(c, &account)
	if err != nil {
		log.Printf("Account error: %v\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	uid := account.Uid

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all documents for the user
	cursor, err := db.Collection.Find(ctx, bson.M{"uid": uid, "record.rid": bson.M{"$exists": true}})
	if err != nil {
		log.Printf("Find error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch records"})
		return
	}
	defer cursor.Close(ctx)

	var records []models.RecordList
	if err = cursor.All(ctx, &records); err != nil {
		log.Printf("Cursor error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode records"})
		return
	}

	if len(records) == 0 {
		log.Printf("No records found for user %s\n", uid)
		c.JSON(http.StatusNotFound, gin.H{"message": "No records found for this user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}

func GetRecordInfo(c *gin.Context) {
	rid := c.Param("rid")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the record with the given rid
	var record models.DBRequest
	err := db.Collection.FindOne(ctx, bson.M{"record.rid": rid}).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Record not found: %s\n", rid)
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		} else {
			log.Printf("Find error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch record"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"record": record})
}

func GetProductInfo(c *gin.Context) {
	var req models.GetProductInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the record with the given rid
	var record models.DBProduct
	err := db.Collection.FindOne(ctx, bson.M{"record.rid": req.Rid, "product.pname": req.Pname}).Decode(&record)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Record or product not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch record"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"record": record})
}

func UpdateProduct(c *gin.Context) {
	var req models.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First check if the record and product exist
	filter := bson.M{
		"record.rid":    req.Rid,
		"product.pname": req.Pname,
	}

	var existingRecord models.DBRequest
	err := db.Collection.FindOne(ctx, filter).Decode(&existingRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Record or product not found: %v\n", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Record or product not found"})
		} else {
			log.Printf("Find error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch record"})
		}
		return
	}

	totalPrice := 0
	for _, product := range existingRecord.Product {
		if product.Pname == req.Pname {
			totalPrice += (req.NewPrice * req.NewAmount)
		} else {
			totalPrice += (product.Price * product.Amount)
		}
	}

	// Update the price of the specific product
	update := bson.M{
		"$set": bson.M{
			"product.$.price":  req.NewPrice,
			"product.$.amount": req.NewAmount,
			"totalPrice":       totalPrice,
		},
	}

	result, err := db.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Update error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record or product not found"})
		return
	}

	// Fetch updated record
	var updatedRecord models.DBRequest
	err = db.Collection.FindOne(ctx, bson.M{"record.rid": req.Rid}).Decode(&updatedRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product price updated successfully",
		"record":  updatedRecord,
	})
}

func UpdateMart(c *gin.Context) {
	var req models.UpdateMartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First check if the record and product exist
	filter := bson.M{
		"record.rid": req.Rid,
	}

	var existingRecord models.DBRequest
	err := db.Collection.FindOne(ctx, filter).Decode(&existingRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch record"})
		}
		return
	}

	// Update the price of the specific product
	update := bson.M{
		"$set": bson.M{
			"mart.martAddress": req.NewMartAddr,
			"mart.martName":    req.NewMartName,
			"mart.tel":         req.NewMartTel,
		},
	}

	result, err := db.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Update error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update mart"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	// Fetch updated mart
	var updatedRecord models.DBRequest
	err = db.Collection.FindOne(ctx, bson.M{"record.rid": req.Rid}).Decode(&updatedRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product price updated successfully",
		"record":  updatedRecord,
	})
}

func UpdateRecord(c *gin.Context) {
	var req models.UpdateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// First check if the record and product exist
	filter := bson.M{
		"record.rid": req.Rid,
	}

	var existingRecord models.DBRequest
	err := db.Collection.FindOne(ctx, filter).Decode(&existingRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch record"})
		}
		return
	}

	update := bson.M{
		"$set": bson.M{
			"record.rname":     req.NewRname,
			"record.timeStamp": req.NewTime,
		},
	}

	result, err := db.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Update error: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update record"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Record not found"})
		return
	}

	// Fetch updated mart
	var updatedRecord models.DBRequest
	err = db.Collection.FindOne(ctx, bson.M{"record.rid": req.Rid}).Decode(&updatedRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product price updated successfully",
		"record":  updatedRecord,
	})
}
