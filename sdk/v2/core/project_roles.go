package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
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
	*restmachinery.BaseClient
}

// NewProjectRolesClient returns a specialized client for managing Project
// Roles.
func NewProjectRolesClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ProjectRolesClient {
	return &projectRolesClient{
		BaseClient: &restmachinery.BaseClient{
			APIAddress: apiAddress,
			APIToken:   apiToken,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: allowInsecure,
					},
				},
			},
		},
	}
}

func (p *projectRolesClient) Grant(
	ctx context.Context,
	projectID string,
	roleAssignment authx.RoleAssignment,
) error {
	return p.ExecuteRequest(
		ctx,
		restmachinery.OutboundRequest{
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
		restmachinery.OutboundRequest{
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
