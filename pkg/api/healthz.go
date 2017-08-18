package api

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"
)

// Healthz is the gin handler for the GET /healthz endpoint
func Healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
