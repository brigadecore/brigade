package api

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/deis/acid/pkg/storage"
)

// Project creates a new gin handler for the GET /project/:id
// endpoint
func Project(storage storage.Store) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Params.ByName("id")
		proj, err := storage.GetProject(id)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		c.JSON(http.StatusOK, proj)
	}
}

// Build creates a new gin handler for the GET /build/:id endpoint
func Build(storage storage.Store) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Params.ByName("id")
		build, err := storage.GetBuild(id)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		c.JSON(http.StatusOK, build)
	}
}
