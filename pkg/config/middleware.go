package config

import (
	"os"

	"gopkg.in/gin-gonic/gin.v1"
)

const ConfigAcidNamespace = "ACID_NAMESPACE"

// Middleware is Gin middleware for injecting configuration into the gin.Context at runtime.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ConfigAcidNamespace, os.Getenv(ConfigAcidNamespace))
		c.Next()
	}
}

// AcidNamespace returns the configured (namespace, true), or ("default", false)
func AcidNamespace(c *gin.Context) (string, bool) {
	val, ok := c.Get(ConfigAcidNamespace)
	if ok {
		return val.(string), ok
	}
	return "default", false
}
