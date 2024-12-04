package handlers

import (
	"context"
	jwt "loginserver/auth"
	"loginserver/db"
	"loginserver/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func ProfileHandler(c *gin.Context) {
	account, err := jwt.GetAccount(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
		return
	}

	email := account.Email

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.DBRequest
	err = db.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uid":      user.Uid,
		"email":    user.Email,
		"nickname": user.Nickname,
	})
}
