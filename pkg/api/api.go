package api

import (
	"io"
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/deis/acid/pkg/storage"
)

// Project creates a new gin handler for the GET /project/:id endpoint
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

// BuildJobs creates a new gin handler for the GET /build/:id/jobs endpoint
func BuildJobs(storage storage.Store) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Params.ByName("id")
		build, err := storage.GetBuild(id)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		jobs, err := storage.GetBuildJobs(build)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		c.JSON(http.StatusOK, jobs)
	}
}

// Job creates a new gin handler for the GET /job/:id endpoint
func Job(storage storage.Store) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Params.ByName("id")
		job, err := storage.GetJob(id)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		c.JSON(http.StatusOK, job)
	}
}

// JobLogs creates a new gin handler for the GET /job/:id/logs endpoint
func JobLogs(storage storage.Store) func(c *gin.Context) {
	return func(c *gin.Context) {
		id := c.Params.ByName("id")
		job, err := storage.GetJob(id)
		if err != nil {
			c.JSON(http.StatusNotFound, struct{}{})
			return
		}
		if c.Query("stream") == "true" {
			logReader, err := storage.GetJobLogStream(job)
			if err != nil {
				c.JSON(http.StatusNotFound, struct{}{})
				return
			}
			defer logReader.Close()
			io.Copy(c.Writer, logReader)
		} else {
			logs, err := storage.GetJobLog(job)
			if err != nil {
				c.JSON(http.StatusNotFound, struct{}{})
				return
			}
			c.JSON(http.StatusOK, logs)
		}
	}
}
