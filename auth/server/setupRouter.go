package server

import (
	"net/http"

	"github.com/SeoPPak/FullStackProgrammingBackend/tree/master/auth/route/login"

	"github.com/gin-gonic/gin"
)

func Setup() *gin.Engine {
	r := gin.Default()

	r.GET("/", login.GoogleForm)
	r.GET("/auth/google/login", login.GoogleLoginHandler)
	r.GET("/auth/google/callback", login.GoogleAuthCallback)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	return r
}
