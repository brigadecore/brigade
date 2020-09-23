package core

import (
	"context"
	"fmt"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/authx"
	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// RoleTypeProject represents a project-level Role.
const RoleTypeProject authx.RoleType = "PROJECT"

const (
	// RoleNameProjectAdmin is the name of a project-level Role that enables
	// principals to manage all aspects of a given Project, including the
	// Project's secrets.
	RoleNameProjectAdmin authx.RoleName = "PROJECT_ADMIN"
	// RoleNameProjectDeveloper is the name of a project-level Role that enables
	// principals to update Projects. This Role does NOT enable event creation
	// or secret management.
	RoleNameProjectDeveloper authx.RoleName = "PROJECT_DEVELOPER"
	// RoleNameProjectUser is the name of a project-level Role that enables
	// principals to create and manage Events for a Project.
	RoleNameProjectUser authx.RoleName = "PROJECT_USER"
)

// ProjectRolesClient is the specialized client for managing project-level
// RoleAssignments with the Brigade API.
type ProjectRolesClient interface {

	// TODO: This needs a function for listing available project roles

	// TODO: This needs a function for listing role assignments by project

	// Grant grants the project-level Role specified by the RoleAssignment to the
	// principal also specified by the RoleAssignment.
	Grant(
		ctx context.Context,
		projectID string,
		roleAssignment authx.RoleAssignment,
	) error
	// Revoke revokes the project-level Role specified by the RoleAssignment for
	// principal also specified by the RoleAssignment.
	Revoke(
		ctx context.Context,
		projectID string,
		roleAssignment authx.RoleAssignment,
	) error
}

type projectRolesClient struct {
	*rm.BaseClient
}

// NewProjectRolesClient returns a specialized client for managing Project
// Roles.
func NewProjectRolesClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) ProjectRolesClient {
	return &projectRolesClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (p *projectRolesClient) Grant(
	ctx context.Context,
	projectID string,
	roleAssignment authx.RoleAssignment,
) error {
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodPost,
			Path: fmt.Sprintf(
				"v2/projects/%s/role-assignments",
				projectID,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  roleAssignment,
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectRolesClient) Revoke(
	ctx context.Context,
	projectID string,
	roleAssignment authx.RoleAssignment,
) error {
	queryParams := map[string]string{
		"role":          string(roleAssignment.Role),
		"principalType": string(roleAssignment.PrincipalType),
		"principalID":   roleAssignment.PrincipalID,
	}
	return p.ExecuteRequest(
		ctx,
		rm.OutboundRequest{
			Method: http.MethodDelete,
			Path: fmt.Sprintf(
				"v2/projects/%s/role-assignments",
				projectID,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
}
