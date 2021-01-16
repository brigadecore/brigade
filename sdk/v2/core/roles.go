package core

import (
	"context"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/authx"
	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// RoleTypeProject represents a project-level Role.
const RoleTypeProject authx.RoleType = "PROJECT"

type projectRoleAssignmentsClient struct {
	*rm.BaseClient
}

// NewProjectRoleAssignmentsClient returns a specialized client for managing
// project-level RoleAssignments.
func NewProjectRoleAssignmentsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) authx.RoleAssignmentsClient {
	return &projectRoleAssignmentsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (p *projectRoleAssignmentsClient) Grant(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/project-role-assignments",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  roleAssignment,
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectRoleAssignmentsClient) Revoke(
	ctx context.Context,
	roleAssignment authx.RoleAssignment,
) error {
	queryParams := map[string]string{
		"roleName":      string(roleAssignment.Role.Name),
		"projectID":     roleAssignment.Role.Scope,
		"principalType": string(roleAssignment.Principal.Type),
		"principalID":   roleAssignment.Principal.ID,
	}
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        "v2/project-role-assignments",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
