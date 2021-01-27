package authz

import (
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// APIClient is the root of a tree of more specialized API clients within the
// authz package.
type APIClient interface {
	// RoleAssignments returns a specialized client for managing RoleAssignments.
	RoleAssignments() RoleAssignmentsClient
}

type apiClient struct {
	// roleAssignmentsClient is a specialized client for managing RoleAssignments.
	roleAssignmentsClient RoleAssignmentsClient
}

// NewAPIClient returns an APIClient, which is the root of a tree of more
// specialized API clients within the authz package. It will initialize all
// clients in the tree so they are ready for immediate use.
func NewAPIClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) APIClient {
	return &apiClient{
		roleAssignmentsClient: NewRoleAssignmentsClient(apiAddress, apiToken, opts),
	}
}

func (a *apiClient) RoleAssignments() RoleAssignmentsClient {
	return a.roleAssignmentsClient
}
