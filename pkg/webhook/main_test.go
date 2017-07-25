package webhook

import (
	"os"
	"testing"

	"gopkg.in/gin-gonic/gin.v1"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
