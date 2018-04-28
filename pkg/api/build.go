package api

import (
	"io"
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/Azure/brigade/pkg/storage"
)

// Build represents the build api handlers.
type Build struct {
	store storage.Store
}

// Get creates a new gin handler for the GET /build/:id endpoint
func (api Build) Get(c *gin.Context) {
	id := c.Params.ByName("id")
	// For now, we always get the worker.
	build, err := api.store.GetBuild(id)
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}
	c.JSON(http.StatusOK, build)
}

// Jobs creates a new gin handler for the GET /build/:id/jobs endpoint
func (api Build) Jobs(c *gin.Context) {
	id := c.Params.ByName("id")
	build, err := api.store.GetBuild(id)
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}
	jobs, err := api.store.GetBuildJobs(build)
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// Logs creates a new gin handler for the GET /build/:id/logs endpoint
func (api Build) Logs(c *gin.Context) {
	id := c.Params.ByName("id")
	build, err := api.store.GetBuild(id)
	if err != nil {
		c.JSON(http.StatusNotFound, struct{}{})
		return
	}
	if c.Query("stream") == "true" {
		logReader, err := api.store.GetWorkerLogStream(build.Worker)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		defer logReader.Close()
		io.Copy(c.Writer, logReader)
	} else {
		logs, err := api.store.GetWorkerLog(build.Worker)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		if len(logs) == 0 {
			c.JSON(http.StatusNoContent, nil)
		}
		c.JSON(http.StatusOK, logs)
	}
}
