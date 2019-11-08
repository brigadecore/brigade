package api

import (
	"io"
	"net/http"
	"sort"

	restful "github.com/emicklei/go-restful"

	"github.com/brigadecore/brigade/pkg/storage"
)

// Build represents the build api handlers.
type Build struct {
	store storage.Store
}

// Get creates a new gin handler for the GET /build/:id endpoint
func (api Build) Get(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	// For now, we always get the worker.
	build, err := api.store.GetBuild(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "Build could not be found.")
		return
	}
	response.WriteEntity(build)
}

// Jobs creates a new gin handler for the GET /build/:id/jobs endpoint
func (api Build) Jobs(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	build, err := api.store.GetBuild(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "Build could not be found.")
		return
	}
	jobs, err := api.store.GetBuildJobs(build)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "Build Jobs could not be found.")
		return
	}
	sort.Slice(jobs, func(i, j int) bool { return jobs[i].CreationTime.String() < jobs[j].CreationTime.String() })
	response.WriteEntity(jobs)

}

// Logs creates a new gin handler for the GET /build/:id/logs endpoint
func (api Build) Logs(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	build, err := api.store.GetBuild(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "Build could not be found.")
		return
	}
	if request.QueryParameter("stream") == "true" {
		logReader, err := api.store.GetWorkerLogStream(build.Worker)
		if err != nil {
			response.WriteErrorString(http.StatusNotFound, "Build Logs could not be found.")
			return
		}
		defer logReader.Close()
		io.Copy(response.ResponseWriter, logReader)
	} else {
		logs, err := api.store.GetWorkerLog(build.Worker)
		if err != nil {
			response.WriteErrorString(http.StatusNotFound, "Build Logs could not be found.")
			return
		}
		if len(logs) == 0 {
			response.WriteErrorString(http.StatusNoContent, "Build Logs Empty")
		}
		response.WriteEntity(logs)
	}
}
