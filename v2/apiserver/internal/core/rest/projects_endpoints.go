package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

// ProjectsEndpoints implements restmachinery.Endpoints to provide
// Project-related URL --> action mappings to a restmachinery.Server.
type ProjectsEndpoints struct {
	AuthFilter          restmachinery.Filter
	ProjectSchemaLoader gojsonschema.JSONLoader
	Service             core.ProjectsService
}

// Register is invoked by restmachinery.Server to register Project-related
// URL --> action mappings to a restmachinery.Server.
func (p *ProjectsEndpoints) Register(router *mux.Router) {
	// Create Project
	router.HandleFunc(
		"/v2/projects",
		p.AuthFilter.Decorate(p.create),
	).Methods(http.MethodPost)

	// List Projects
	router.HandleFunc(
		"/v2/projects",
		p.AuthFilter.Decorate(p.list),
	).Methods(http.MethodGet)

	// Get Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.AuthFilter.Decorate(p.get),
	).Methods(http.MethodGet)

	// Update Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.AuthFilter.Decorate(p.update),
	).Methods(http.MethodPut)

	// Delete Project
	router.HandleFunc(
		"/v2/projects/{id}",
		p.AuthFilter.Decorate(p.delete),
	).Methods(http.MethodDelete)
}

func (p *ProjectsEndpoints) create(w http.ResponseWriter, r *http.Request) {
	project := core.Project{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.ProjectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				return p.Service.Create(r.Context(), project)
			},
			SuccessCode: http.StatusCreated,
		},
	)
}

func (p *ProjectsEndpoints) list(w http.ResponseWriter, r *http.Request) {
	opts := meta.ListOptions{
		Continue: r.URL.Query().Get("continue"),
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		var err error
		if opts.Limit, err = strconv.ParseInt(limitStr, 10, 64); err != nil ||
			opts.Limit < 1 || opts.Limit > 100 {
			restmachinery.WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&meta.ErrBadRequest{
					Reason: fmt.Sprintf(
						`Invalid value %q for "limit" query parameter`,
						limitStr,
					),
				},
			)
			return
		}
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return p.Service.List(r.Context(), opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *ProjectsEndpoints) get(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return p.Service.Get(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *ProjectsEndpoints) update(w http.ResponseWriter, r *http.Request) {
	project := core.Project{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.ProjectSchemaLoader,
			ReqBodyObj:          &project,
			EndpointLogic: func() (interface{}, error) {
				if mux.Vars(r)["id"] != project.ID {
					return nil, &meta.ErrBadRequest{
						Reason: "The project IDs in the URL path and request body do " +
							"not match.",
					}
				}
				return project, p.Service.Update(r.Context(), project)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *ProjectsEndpoints) delete(w http.ResponseWriter, r *http.Request) {
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.Service.Delete(r.Context(), mux.Vars(r)["id"])
			},
			SuccessCode: http.StatusOK,
		},
	)
}
