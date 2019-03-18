package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	restful "github.com/emicklei/go-restful"

	"github.com/brigadecore/brigade/pkg/storage/mock"
)

func TestBuildLogs(t *testing.T) {
	store := mock.New()
	mockAPI := New(store)

	//ctx, _ := gin.CreateTestContext(rw)

	// There is a bug in Gin that will cause a panic if we don't send a request
	// that has a query param.

	httpRequest, _ := http.NewRequest("GET", "/?foo=bar", bytes.NewBuffer(nil))
	req := restful.NewRequest(httpRequest)
	httpWriter := httptest.NewRecorder()
	respo := restful.NewResponse(httpWriter)
	respo.SetRequestAccepts("application/json")

	mockAPI.Build().Logs(req, respo)
	logLines := httpWriter.Body.String()
	expect := fmt.Sprintf("%q", mock.StubLogData)
	if logLines != expect {
		t.Errorf("Expected %q, got %q", expect, logLines)
	}

	// Retest with streaming on, which should return line data instead of JSON data.
	httpWriter = httptest.NewRecorder()
	httpRequest = httptest.NewRequest("GET", "/?stream=true", bytes.NewBuffer(nil))
	respo = restful.NewResponse(httpWriter)
	req = restful.NewRequest(httpRequest)

	mockAPI.Build().Logs(req, respo)
	logLines = httpWriter.Body.String()
	if logLines != mock.StubLogData {
		t.Errorf("Expected %q, got %q", mock.StubLogData, logLines)
	}

	// Check that we get a 204 for no content.
	// Retest with streaming on, which should return line data instead of JSON data.
	store.LogData = ""
	httpWriter = httptest.NewRecorder()
	respo = restful.NewResponse(httpWriter)
	httpRequest = httptest.NewRequest("GET", "/?a=b", bytes.NewBuffer(nil))
	req = restful.NewRequest(httpRequest)

	mockAPI.Build().Logs(req, respo)
	if httpWriter.Code != 204 {
		t.Errorf("Expected %q, got %q", mock.StubLogData, logLines)
	}

}
