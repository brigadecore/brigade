package system

import "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"

// APIClient is the root of a tree of more specialized API clients within the
// system package.
type APIClient interface {
	// Roles returns a specialized client for system Role management.
	Roles() RolesClient
}

type apiClient struct {
	// rolesClient is a specialized client for system Role management.
	rolesClient RolesClient
}

// NewAPIClient returns an APIClient, which is the root of a tree of more
// specialized API clients within the system package. It will initialize all
// clients in the tree so they are ready for immediate use.
func NewAPIClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) APIClient {
	return &apiClient{
		rolesClient: NewRolesClient(apiAddress, apiToken, opts),
	}
}

func (a *apiClient) Roles() RolesClient {
	return a.rolesClient
}
