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

// RoleAssignmentsEndpoints implements restmachinery.Endpoints to provide
// RoleAssignments-related URL --> action mappings to a restmachinery.Server.
type RoleAssignmentsEndpoints struct {
	AuthFilter                 restmachinery.Filter
	RoleAssignmentSchemaLoader gojsonschema.JSONLoader
	Service                    api.RoleAssignmentsService
}

func (r *RoleAssignmentsEndpoints) Register(router *mux.Router) {
	// Grant a Role to a User or Service Account
	router.HandleFunc(
		"/v2/role-assignments",
		r.AuthFilter.Decorate(r.grant),
	).Methods(http.MethodPost)

	// List associations between Users and/or Service Accounts and Roles
	router.HandleFunc(
		"/v2/role-assignments",
		r.AuthFilter.Decorate(r.list),
	).Methods(http.MethodGet)

	// Revoke a Role for a User or Service Account
	router.HandleFunc(
		"/v2/role-assignments",
		r.AuthFilter.Decorate(r.revoke),
	).Methods(http.MethodDelete)
}

func (r *RoleAssignmentsEndpoints) grant(
	w http.ResponseWriter,
	req *http.Request,
) {
	roleAssignment := api.RoleAssignment{}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W:                   w,
			R:                   req,
			ReqBodySchemaLoader: r.RoleAssignmentSchemaLoader,
			ReqBodyObj:          &roleAssignment,
			EndpointLogic: func() (interface{}, error) {
				return nil, r.Service.Grant(req.Context(), roleAssignment)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (r *RoleAssignmentsEndpoints) list(
	w http.ResponseWriter,
	req *http.Request,
) {
	principalType := req.URL.Query().Get("principalType")
	principalID := req.URL.Query().Get("principalID")
	selector := api.RoleAssignmentsSelector{
		Role: api.Role(req.URL.Query().Get("role")),
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
				return r.Service.List(req.Context(), selector, opts)
			},
			SuccessCode: http.StatusOK,
		},
	)
}

func (r *RoleAssignmentsEndpoints) revoke(
	w http.ResponseWriter,
	req *http.Request,
) {
	roleAssignment := api.RoleAssignment{
		Role:  api.Role(req.URL.Query().Get("role")),
		Scope: req.URL.Query().Get("scope"),
		Principal: api.PrincipalReference{
			Type: api.PrincipalType(req.URL.Query().Get("principalType")),
			ID:   req.URL.Query().Get("principalID"),
		},
	}
	restmachinery.ServeRequest(
		restmachinery.InboundRequest{
			W: w,
			R: req,
			EndpointLogic: func() (interface{}, error) {
				return nil, r.Service.Revoke(req.Context(), roleAssignment)
			},
			SuccessCode: http.StatusOK,
		},
	)
}
