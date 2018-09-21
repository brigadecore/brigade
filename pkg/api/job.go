package api

import (
	"io"
	"net/http"

	"github.com/Azure/brigade/pkg/storage"

	restful "github.com/emicklei/go-restful"
)

// Job represents the job api handlers.
type Job struct {
	store storage.Store
}

// Get creates a new gin handler for the GET /job/:id endpoint
func (api Job) Get(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	job, err := api.store.GetJob(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "Job could not be found.")
		return
	}
	response.WriteEntity(job)
}

// Logs creates a new gin handler for the GET /job/:id/logs endpoint
func (api Job) Logs(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	job, err := api.store.GetJob(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "Job could not be found.")
		return
	}
	if request.QueryParameter("stream") == "true" {
		logReader, err := api.store.GetJobLogStream(job)
		if err != nil {
			response.WriteErrorString(http.StatusNotFound, "Job could not be found.")
			return
		}
		defer logReader.Close()
		io.Copy(response.ResponseWriter, logReader)
	} else {
		logs, err := api.store.GetJobLog(job)
		if err != nil {
			response.WriteErrorString(http.StatusNotFound, "Job Logs could not be found.")
			return
		}
		if len(logs) == 0 {
			response.WriteErrorString(http.StatusNoContent, "Job Logs Empty")
		}
		response.WriteEntity(logs)
	}
}
