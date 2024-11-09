package handlers

import (
	"context"
	"net/http"
	"server/db"
	"server/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/bson"
)

func ProfileHandler(c *gin.Context) {
	session := c.MustGet("session").(*sessions.Session)
	email := session.Values["email"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.DBRequest
	err := db.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":    user.Email,
		"nickname": user.Nickname,
	})
}
