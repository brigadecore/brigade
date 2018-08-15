package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/emicklei/go-restful"
)

func dummy(req *restful.Request, resp *restful.Response) { io.WriteString(resp.ResponseWriter, "dummy") }

func TestDefaultCORSFilter(t *testing.T) {
	container := restful.NewContainer()
	container.ServeMux = http.DefaultServeMux
	cors := DefaultCORS
	cors.Container = container
	container.Filter(cors.Filter)

	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(dummy))
	container.Add(ws)

	httpRequest, _ := http.NewRequest("GET", "http://api.alice.com/", nil)
	httpRequest.Header.Set(restful.HEADER_Origin, "http://api.bob.com")

	httpWriter := httptest.NewRecorder()
	container.Dispatch(httpWriter, httpRequest)

	actual := httpWriter.Header().Get(restful.HEADER_AccessControlAllowOrigin)
	if "http://api.bob.com" != actual {
		t.Fatalf("expected header '%s' to contain 'http://api.bob.com' but got '%s'", restful.HEADER_AccessControlAllowOrigin, actual)
	}

	if httpWriter.Body.String() != "dummy" {
		t.Fatalf("expected: dummy but got '%s'", httpWriter.Body.String())
	}
}
