package main

import (
	"dbserver/config"
	"dbserver/db"
	login "dbserver/handlers"
	"dbserver/handlers/middleware"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// User represents the user data we'll store in session
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func Setup() *gin.Engine {
	config.Init()
	db.DBInit()

	r := gin.Default()

	config := cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	r.Use(cors.New(config))

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/records", login.SearchByUid)
		protected.GET("/records/:rid", login.GetRecordInfo)
		protected.GET("/records/product/:pid", login.GetProductInfo)

		protected.PUT("/records/update/product", login.UpdateProduct)
		protected.PUT("/records/update/mart", login.UpdateMart)
		protected.PUT("/records/update/record", login.UpdateRecord)
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	return r
}
