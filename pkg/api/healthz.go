package api

import (
	"net/http"

	restful "github.com/emicklei/go-restful"
)

// Healthz is the gin handler for the GET /healthz endpoint
func Healthz(request *restful.Request, response *restful.Response) {
	response.WriteHeaderAndEntity(http.StatusOK, http.StatusText(http.StatusOK))
}
