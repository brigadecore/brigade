package api

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/brigadecore/brigade/pkg/storage/mock"

	restful "github.com/emicklei/go-restful"
)

func TestJobLogs(t *testing.T) {
	store := mock.New()
	mockAPI := New(store)
	rw := httptest.NewRecorder()

	httpRequest := httptest.NewRequest("GET", "/?foo=bar", bytes.NewBuffer(nil))
	req := restful.NewRequest(httpRequest)

	respo := restful.NewResponse(rw)
	respo.SetRequestAccepts("application/json")

	mockAPI.Job().Logs(req, respo)
	logLines := rw.Body.String()
	expect := fmt.Sprintf("%q", mock.StubLogData)
	if logLines != expect {
		t.Errorf("Expected %q, got %q", expect, logLines)
	}

	// Retest with streaming on, which should return line data instead of JSON data.
	rw = httptest.NewRecorder()
	httpRequest = httptest.NewRequest("GET", "/?stream=true", bytes.NewBuffer(nil))
	req = restful.NewRequest(httpRequest)

	respo = restful.NewResponse(rw)

	mockAPI.Job().Logs(req, respo)
	logLines = rw.Body.String()
	if logLines != mock.StubLogData {
		t.Errorf("Expected %q, got %q", mock.StubLogData, logLines)
	}

	// Check that we get a 204 for no content.
	// Retest with streaming on, which should return line data instead of JSON data.
	store.LogData = ""
	rw = httptest.NewRecorder()
	httpRequest = httptest.NewRequest("GET", "/?a=b", bytes.NewBuffer(nil))
	req = restful.NewRequest(httpRequest)

	respo = restful.NewResponse(rw)

	mockAPI.Job().Logs(req, respo)
	if rw.Code != 204 {
		t.Errorf("Expected %q, got %q", mock.StubLogData, logLines)
	}

}
