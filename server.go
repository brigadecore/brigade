package main

import (
	"net/http"

	"github.com/deis/acid/pkg/config"
	"github.com/deis/acid/pkg/storage"
	"github.com/deis/acid/pkg/webhook"

	"gopkg.in/gin-gonic/gin.v1"
)

func main() {
	router := gin.New()
	router.Use(gin.Recovery())

	store := storage.New()

	events := router.Group("/events")
	{
		events.Use(gin.Logger())
		events.Use(config.Middleware())

		events.POST("/github", webhook.NewGithubHook(store).Handle)
		events.POST("/exec/:org/:project/:commit", webhook.NewExecHook(store).Handle)
	}

	// Lame UI
	ui := router.Group("/log/:org/:project")
	{
		ui.Use(gin.Logger())
		ui.Use(config.Middleware())

		ui.GET("/", logToHTML)
		ui.GET("/status.svg", badge)
		ui.GET("/id/:commit", logToHTML)
	}

	router.GET("/healthz", healthz)

	router.Run(":7744")
}

func healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
