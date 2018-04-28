package api

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
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

// ProjectBuildSummary is a project plus the latest build data
type ProjectBuildSummary struct {
	Project   *brigade.Project `json:"project"`
	LastBuild *brigade.Build   `json:"lastBuild"`
}

// ListWithLatestBuild lists the projects with the latest builds attached.
func (api Project) ListWithLatestBuild(c *gin.Context) {

	projects, err := api.store.GetProjects()
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}

	res := []*ProjectBuildSummary{}
	for _, p := range projects {
		pbs := &ProjectBuildSummary{Project: p}
		builds, err := api.store.GetProjectBuilds(p)
		if err == nil && len(builds) > 0 {
			pbs.LastBuild = builds[len(builds)-1]
		}
		res = append(res, pbs)
	}
	c.JSON(http.StatusOK, res)
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
