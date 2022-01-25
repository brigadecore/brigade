package sdk

import (
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
)

// SystemAuthzClient is the root of a tree of more specialized API clients
// for handling system-wide authorization concerns.
type SystemAuthzClient interface {
	// RoleAssignments returns a specialized client for managing RoleAssignments.
	RoleAssignments() RoleAssignmentsClient
}

type systemAuthzClient struct {
	// roleAssignmentsClient is a specialized client for managing RoleAssignments.
	roleAssignmentsClient RoleAssignmentsClient
}

// NewSystemAuthzClient returns an SystemAuthzClient, which is the root of a
// tree of more specialized API clients for handling system-wide authorization
// concerns. It will initialize all clients in the tree so they are ready for
// immediate use.
func NewSystemAuthzClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) SystemAuthzClient {
	return &systemAuthzClient{
		roleAssignmentsClient: NewRoleAssignmentsClient(apiAddress, apiToken, opts),
	}
}

func (s *systemAuthzClient) RoleAssignments() RoleAssignmentsClient {
	return s.roleAssignmentsClient
}
