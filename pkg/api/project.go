package api

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/deis/brigade/pkg/storage"
)

// Project represents the project api handlers.
type Project struct {
	store storage.Store
}

// List creates a new gin handler for the GET /projects endpoint
func (api Project) List(c *gin.Context) {
	projects, err := api.store.GetProjects()
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}
	c.JSON(http.StatusOK, projects)
}

// Get creates a new gin handler for the GET /project/:id endpoint
func (api Project) Get(c *gin.Context) {
	id := c.Params.ByName("id")
	proj, err := api.store.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}
	c.JSON(http.StatusOK, proj)
}

// Builds creates a new gin handler for the GET /project/:id/builds endpoint
func (api Project) Builds(c *gin.Context) {
	id := c.Params.ByName("id")
	proj, err := api.store.GetProject(id)
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}
	builds, err := api.store.GetProjectBuilds(proj)
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}
	c.JSON(http.StatusOK, builds)
}
