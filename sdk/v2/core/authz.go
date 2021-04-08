package core

import (
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// AuthzClient is the specialized client for managing project-level
// authorization concerns with the Brigade API.
type AuthzClient interface {
	// ProjectRoleAssignments returns a specialized client for managing
	// ProjectRoleAssignments.
	ProjectRoleAssignments() ProjectRoleAssignmentsClient
}

type authzClient struct {
	// projectRoleAssignmentsClient is a specialized client for managing
	// ProjectRoleAssignments.
	projectRoleAssignmentsClient ProjectRoleAssignmentsClient
}

// NewAuthzClient returns a specialized client for managing project-level
// authorization concerns.
func NewAuthzClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) AuthzClient {
	return &authzClient{
		projectRoleAssignmentsClient: NewProjectRoleAssignmentsClient(
			apiAddress,
			apiToken,
			opts,
		),
	}
}

func (a *authzClient) ProjectRoleAssignments() ProjectRoleAssignmentsClient {
	return a.projectRoleAssignmentsClient
}
