package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/v2/apiserver/internal/api"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

// ProjectRoleAssignmentsEndpoints implements restmachinery.Endpoints to provide
// ProjectRoleAssignment-related URL --> action mappings to a
// restmachinery.Server.
type ProjectRoleAssignmentsEndpoints struct {
	AuthFilter                        restmachinery.Filter
	ProjectRoleAssignmentSchemaLoader gojsonschema.JSONLoader
	Service                           api.ProjectRoleAssignmentsService
}

func (p *ProjectRoleAssignmentsEndpoints) Register(router *mux.Router) {
	// Grant a Project Role to a User or Service Account
	router.HandleFunc(
		"/v2/projects/{id}/role-assignments",
		p.AuthFilter.Decorate(p.grant),
	).Methods(http.MethodPost)

	// List associations between Users and/or Service Accounts and Roles
	router.HandleFunc(
		"/v2/projects/{id}/role-assignments",
		p.AuthFilter.Decorate(p.list),
	).Methods(http.MethodGet)

	// Newer role-assignments endpoint that supports querying across projects
	router.HandleFunc(
		"/v2/project-role-assignments",
		p.AuthFilter.Decorate(p.list),
	).Methods(http.MethodGet)

	// Revoke a Project Role for a User or Service Account
	router.HandleFunc(
		"/v2/projects/{id}/role-assignments",
		p.AuthFilter.Decorate(p.revoke),
	).Methods(http.MethodDelete)
}

func (p *ProjectRoleAssignmentsEndpoints) grant(
	w http.ResponseWriter,
	r *http.Request,
) {
	projectRoleAssignment := api.ProjectRoleAssignment{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.ProjectRoleAssignmentSchemaLoader,
			ReqBodyObj:          &projectRoleAssignment,
			EndpointLogic: func() (interface{}, error) {
				projectRoleAssignment.ProjectID = mux.Vars(r)["id"]
				return nil, p.Service.Grant(r.Context(), projectRoleAssignment)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *ProjectRoleAssignmentsEndpoints) list(
	w http.ResponseWriter,
	req *http.Request,
) {
	principalType := req.URL.Query().Get("principalType")
	principalID := req.URL.Query().Get("principalID")

	// This will yield a project ID if we got here via the original
	// /v2/projects/{id}/role-assignments endpoint.
	projectID := mux.Vars(req)["id"]
	// If we cannot pick a project ID out of the path, try to get it from a query
	// parameter.
	if projectID == "" {
		projectID = req.URL.Query().Get("project")
	}
	// It's possible and legitimate if projectID is still an empty-string at this
	// point.

	selector := api.ProjectRoleAssignmentsSelector{
		ProjectID: projectID,
		Role:      api.Role(req.URL.Query().Get("role")),
	}
	if principalType != "" && principalID != "" {
		selector.Principal = &api.PrincipalReference{
			Type: api.PrincipalType(req.URL.Query().Get("principalType")),
			ID:   req.URL.Query().Get("principalID"),
		}
	}
	opts := meta.ListOptions{
		Continue: req.URL.Query().Get("continue"),
	}
	if limitStr := req.URL.Query().Get("limit"); limitStr != "" {
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
			R: req,
			EndpointLogic: func() (interface{}, error) {
				return p.Service.List(req.Context(), selector, opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *ProjectRoleAssignmentsEndpoints) revoke(
	w http.ResponseWriter,
	r *http.Request,
) {
	projectRoleAssignment := api.ProjectRoleAssignment{
		ProjectID: mux.Vars(r)["id"],
		Role:      api.Role(r.URL.Query().Get("role")),
		Principal: api.PrincipalReference{
			Type: api.PrincipalType(r.URL.Query().Get("principalType")),
			ID:   r.URL.Query().Get("principalID"),
		},
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.Service.Revoke(r.Context(), projectRoleAssignment)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
