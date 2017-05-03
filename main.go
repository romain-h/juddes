package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/romain-h/juddes/gists"
)

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	r.GET("/sync-all", func(c *gin.Context) {
		gists.Sync()
	})

	r.Run(":" + os.Getenv("PORT"))
}
