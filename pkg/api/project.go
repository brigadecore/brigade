package api

import (
	"net/http"

	"github.com/deis/acid/pkg/config"
	"github.com/deis/acid/pkg/storage"
	"gopkg.in/gin-gonic/gin.v1"
)

// Project creates a new gin handler for the GET /project/:id
// endpoint
func Project(storage storage.Store) func(c *gin.Context) {
	return func(c *gin.Context) {
		namespace, _ := config.AcidNamespace(c)
		id := c.Params.ByName("id")
		proj, err := storage.GetProject(id, namespace)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		c.JSON(http.StatusOK, proj)
	}
}
