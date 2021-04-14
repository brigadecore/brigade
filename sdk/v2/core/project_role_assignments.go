package core

import (
	"context"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/authz"
	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	libAuthz "github.com/brigadecore/brigade/sdk/v2/lib/authz"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

type projectRoleAssignmentsClient struct {
	*rm.BaseClient
}

// NewProjectRoleAssignmentsClient returns a specialized client for managing
// project-level RoleAssignments.
func NewProjectRoleAssignmentsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) authz.RoleAssignmentsClient {
	return &projectRoleAssignmentsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (p *projectRoleAssignmentsClient) Grant(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/project-role-assignments",
			ReqBodyObj:  roleAssignment,
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectRoleAssignmentsClient) Revoke(
	ctx context.Context,
	roleAssignment libAuthz.RoleAssignment,
) error {
	queryParams := map[string]string{
		"roleName":      string(roleAssignment.Role.Name),
		"roleScope":     roleAssignment.Role.Scope,
		"principalType": string(roleAssignment.Principal.Type),
		"principalID":   roleAssignment.Principal.ID,
	}
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/project-role-assignments",
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
