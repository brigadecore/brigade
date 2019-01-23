package api

import (
	"net/http"
	"sort"

	restful "github.com/emicklei/go-restful"

	"github.com/Azure/brigade/pkg/brigade"
	"github.com/Azure/brigade/pkg/storage"
)

// Project represents the project api handlers.
type Project struct {
	store storage.Store
}

// List creates a new gin handler for the GET /projects endpoint
func (api Project) List(request *restful.Request, response *restful.Response) {
	projects, err := api.store.GetProjects()
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "No Projects found.")
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, projects)
}

// ProjectBuildSummary is a project plus the latest build data
type ProjectBuildSummary struct {
	Project   *brigade.Project `json:"project"`
	LastBuild *brigade.Build   `json:"lastBuild"`
}

// ListWithLatestBuild lists the projects with the latest builds attached.
func (api Project) ListWithLatestBuild(request *restful.Request, response *restful.Response) {
	projects, err := api.store.GetProjects()
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "No Projects found.")
		return
	}
	res := api.getBuildSummariesForProjects(projects)

	response.WriteHeaderAndEntity(http.StatusOK, res)
}

func (api Project) getBuildSummariesForProjects(projects []*brigade.Project) []*ProjectBuildSummary {
	res := []*ProjectBuildSummary{}
	for _, p := range projects {
		pbs := &ProjectBuildSummary{Project: p}
		builds, err := api.store.GetProjectBuilds(p)
		if err == nil && len(builds) > 0 {
			sort.Slice(builds, func(i, j int) bool {
				if builds[i].Worker == nil || builds[j].Worker == nil {
					return false
				}
				return builds[i].Worker.StartTime.Before(builds[j].Worker.StartTime)
			})
			pbs.LastBuild = builds[len(builds)-1]
		}
		res = append(res, pbs)
	}
	return res
}

// Get creates a new gin handler for the GET /project/:id endpoint
func (api Project) Get(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	proj, err := api.store.GetProject(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "No Project found.")
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, proj)
}

// Builds creates a new gin handler for the GET /project/:id/builds endpoint
func (api Project) Builds(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("id")
	proj, err := api.store.GetProject(id)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "No Project found.")
		return
	}
	builds, err := api.store.GetProjectBuilds(proj)
	if err != nil {
		response.WriteErrorString(http.StatusNotFound, "No Project Builds found.")
		return
	}
	response.WriteHeaderAndEntity(http.StatusOK, builds)
}
