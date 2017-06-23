package main

import (
	"net/http"

	"github.com/deis/acid/pkg/webhook"

	"gopkg.in/gin-gonic/gin.v1"
)

func main() {
	router := gin.New()
	router.Use(gin.Recovery())

	events := router.Group("/events", gin.Logger())
	{
		events.POST("/github", webhook.EventRouter)
	}

	// Lame UI
	logs := router.Group("/log/:org/:project", gin.Logger())
	{
		logs.GET("/", logToHTML)
		logs.GET("/status.svg", badge)
		logs.GET("/id/:commit", logToHTML)
	}

	router.GET("/healthz", healthz)

	router.Run(":7744")
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
