package webhook

import (
	"net/http"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/Azure/brigade/pkg/storage"
)

func Activate(store storage.Store, gatewayHost string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if gatewayHost == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "No host is configured."})
			return
		}

		proj, err := store.GetProject(c.Param("project"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := CreateHook(proj, gatewayHost); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
	}
}
