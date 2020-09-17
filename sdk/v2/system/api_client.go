package system

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
func NewAPIClient(apiAddress, apiToken string, allowInsecure bool) APIClient {
	return &apiClient{
		rolesClient: NewRolesClient(apiAddress, apiToken, allowInsecure),
	}
}

func (a *apiClient) Roles() RolesClient {
	return a.rolesClient
}
