package api

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"gopkg.in/gin-gonic/gin.v1"

	"github.com/Azure/brigade/pkg/storage/mock"
)

func TestBuildLogs(t *testing.T) {
	store := mock.New()
	mockAPI := New(store)
	rw := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rw)

	// There is a bug in Gin that will cause a panic if we don't send a request
	// that has a query param.
	ctx.Request = httptest.NewRequest("GET", "/?foo=bar", bytes.NewBuffer(nil))

	mockAPI.Build().Logs(ctx)
	logLines := rw.Body.String()
	expect := fmt.Sprintf("%q", mock.StubLogData)
	if logLines != expect {
		t.Errorf("Expected %q, got %q", expect, logLines)
	}

	// Retest with streaming on, which should return line data instead of JSON data.
	rw = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(rw)
	ctx.Request = httptest.NewRequest("GET", "/?stream=true", bytes.NewBuffer(nil))

	mockAPI.Build().Logs(ctx)
	logLines = rw.Body.String()
	if logLines != mock.StubLogData {
		t.Errorf("Expected %q, got %q", mock.StubLogData, logLines)
	}

	// Check that we get a 204 for no content.
	// Retest with streaming on, which should return line data instead of JSON data.
	store.LogData = ""
	rw = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(rw)
	ctx.Request = httptest.NewRequest("GET", "/?a=b", bytes.NewBuffer(nil))

	mockAPI.Build().Logs(ctx)
	if rw.Code != 204 {
		t.Errorf("Expected %q, got %q", mock.StubLogData, logLines)
	}

}
