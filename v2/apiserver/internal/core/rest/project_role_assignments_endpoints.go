package rest

import (
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/restmachinery"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"
)

// ProjectRoleAssignmentsEndpoints implements restmachinery.Endpoints to provide
// project-level RoleAssignment-related URL --> action mappings to a
// restmachinery.Server.
type ProjectRoleAssignmentsEndpoints struct {
	AuthFilter                        restmachinery.Filter
	ProjectRoleAssignmentSchemaLoader gojsonschema.JSONLoader
	Service                           authz.RoleAssignmentsService
}

func (p *ProjectRoleAssignmentsEndpoints) Register(router *mux.Router) {
	// Grant a Project Role to a User or Service Account
	router.HandleFunc(
		"/v2/project-role-assignments",
		p.AuthFilter.Decorate(p.grant),
	).Methods(http.MethodPost)

	// Revoke a Project Role for a User or Service Account
	router.HandleFunc(
		"/v2/project-role-assignments",
		p.AuthFilter.Decorate(p.revoke),
	).Methods(http.MethodDelete)
}

func (p *ProjectRoleAssignmentsEndpoints) grant(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := libAuthz.RoleAssignment{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   r,
			ReqBodySchemaLoader: p.ProjectRoleAssignmentSchemaLoader,
			ReqBodyObj:          &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.Service.Grant(r.Context(), roleAssignment)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *ProjectRoleAssignmentsEndpoints) revoke(
	w http.ResponseWriter,
	r *http.Request,
) {
	roleAssignment := libAuthz.RoleAssignment{
		Role: libAuthz.Role{
			Type: core.RoleTypeProject,
			Name: libAuthz.RoleName(r.URL.Query().Get("role")),
		},
		Scope: r.URL.Query().Get("scope"),
		Principal: libAuthz.PrincipalReference{
			Type: libAuthz.PrincipalType(r.URL.Query().Get("principalType")),
			ID:   r.URL.Query().Get("principalID"),
		},
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: r,
			EndpointLogic: func() (interface{}, error) {
				return nil, p.Service.Revoke(r.Context(), roleAssignment)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
