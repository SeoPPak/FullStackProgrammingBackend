package main

import (
	"loginserver/config"
	"loginserver/db"
	login "loginserver/handlers"
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

	r.GET("/", login.GoogleForm)
	r.GET("/auth/google/login", login.GoogleLoginHandler)
	r.GET("/auth/google/callback", login.GoogleAuthCallback)
	r.POST("/auth/google/verify", login.GoogleTokenVerifyHandler)

	r.POST("/auth/signup", login.HandleSignup)
	r.POST("/auth/login", login.HandleLogin)
	r.GET("/auth/logout", login.HandleLogout)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	return r
}
