package sdk

import (
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
)

// ProjectAuthzClient is the specialized client for managing project-level
// authorization concerns with the Brigade API.
type ProjectAuthzClient interface {
	// RoleAssignments returns a specialized client for managing project-level
	// RoleAssignments.
	RoleAssignments() ProjectRoleAssignmentsClient
}

type projectAuthzClient struct {
	// projectRoleAssignmentsClient is a specialized client for managing
	// ProjectRoleAssignments.
	projectRoleAssignmentsClient ProjectRoleAssignmentsClient
}

// NewProjectAuthzClient returns a specialized client for managing project-level
// authorization concerns.
func NewProjectAuthzClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) ProjectAuthzClient {
	return &projectAuthzClient{
		projectRoleAssignmentsClient: NewProjectRoleAssignmentsClient(
			apiAddress,
			apiToken,
			opts,
		),
	}
}

func (p *projectAuthzClient) RoleAssignments() ProjectRoleAssignmentsClient {
	return p.projectRoleAssignmentsClient
}
