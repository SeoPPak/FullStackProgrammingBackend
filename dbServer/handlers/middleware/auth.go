package middleware

import (
	jwt "dbserver/auth"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c, err := jwt.FillContext(c)
		if err != nil {
			log.Printf("Error: can't fill context: %s\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		log.Printf("auth success\n")
		c.Next()
	}
}
