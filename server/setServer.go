package main

import (
	"net/http"
	"server/config"
	"server/db"
	login "server/handlers"

	"github.com/gorilla/sessions"

	"github.com/gin-gonic/gin"
)

// User represents the user data we'll store in session
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func SessionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := config.AppConfig.SessionStore.Get(c.Request, "session-name")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get session"})
			c.Abort()
			return
		}
		c.Set("session", session)
		c.Next()
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := c.MustGet("session").(*sessions.Session)
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func Setup() *gin.Engine {
	config.Init()
	db.DBInit()

	r := gin.Default()

	r.Use(SessionMiddleware())
	r.GET("/", login.GoogleForm)
	r.GET("/auth/google/login", login.GoogleLoginHandler)
	r.GET("/auth/google/callback", login.GoogleAuthCallback)

	r.POST("/signup", login.HandleSignup)
	r.POST("/auth/login", login.HandleLogin)
	r.GET("/profile", AuthRequired(), login.ProfileHandler)

	r.GET("/ocr/data", login.RequestOCR)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	/*
		protected := r.Group("/")
		protected.Use(AuthRequired()){
			protected.GET("/profile", )
		}
	*/

	return r
}
