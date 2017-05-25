package main

import (
	"github.com/deis/acid/pkg/webhook"

	"gopkg.in/gin-gonic/gin.v1"
)

func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) { c.JSON(200, gin.H{"message": "OK"}) })
	router.POST("/events/github", webhook.EventRouter)

	// Lame UI
	router.GET("/log/:org/:project", logToHTML)

	router.Run(":7744")
}
