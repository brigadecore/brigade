package core

import (
	"github.com/brigadecore/brigade/sdk/v2/authz"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// AuthzClient is the specialized client for managing project-level
// authorization concerns with the Brigade API.
type AuthzClient interface {
	// RoleAssignments returns a specialized client for managing project-level
	// RoleAssignments.
	RoleAssignments() authz.RoleAssignmentsClient
}

type authzClient struct {
	// roleAssignmentsClient is a specialized client for managing project-level
	// RoleAssignments.
	roleAssignmentsClient authz.RoleAssignmentsClient
}

// NewAuthzClient returns a specialized client for managing project-level
// authorization concerns.
func NewAuthzClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) AuthzClient {
	return &authzClient{
		roleAssignmentsClient: NewProjectRoleAssignmentsClient(
			apiAddress,
			apiToken,
			opts,
		),
	}
}

func (a *authzClient) RoleAssignments() authz.RoleAssignmentsClient {
	return a.roleAssignmentsClient
}
