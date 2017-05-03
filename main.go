package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/romain-h/juddes/gists"
)

var authorizationHeader = fmt.Sprintf("Bearer %s", os.Getenv("JUDDES_ACCESS_TOKEN"))

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Authorization") != authorizationHeader {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			c.Abort()
		}
	}
}

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	authorized := r.Group("/", Auth()) // Apply Auth middleware

	authorized.GET("/sync-all", func(c *gin.Context) {
		gists.Sync()
	})

	r.Run(":" + os.Getenv("PORT"))
}
