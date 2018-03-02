package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Healthz is the gin handler for the GET /healthz endpoint
func Healthz(c *gin.Context) {
	c.String(http.StatusOK, http.StatusText(http.StatusOK))
}
