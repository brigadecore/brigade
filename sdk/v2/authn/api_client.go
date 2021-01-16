package authn

import (
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// APIClient is the root of a tree of more specialized API clients within the
// authn package.
type APIClient interface {
	// ServiceAccounts returns a specialized client for ServiceAccount management.
	ServiceAccounts() ServiceAccountsClient
	// Sessions returns a specialized client for Session management.
	Sessions() SessionsClient
	// Users returns a specialized client for User management.
	Users() UsersClient
}

type apiClient struct {
	// serviceAccountsClient is a specialized client for ServiceAccount
	// management.
	serviceAccountsClient ServiceAccountsClient
	// sessionsClient is a specialized client for Session management.
	sessionsClient SessionsClient
	// usersClient is a specialized client for User management.
	usersClient UsersClient
}

// NewAPIClient returns an APIClient, which is the root of a tree of more
// specialized API clients within the authn package. It will initialize all
// clients in the tree so they are ready for immediate use.
func NewAPIClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) APIClient {
	return &apiClient{
		serviceAccountsClient: NewServiceAccountsClient(
			apiAddress,
			apiToken,
			opts,
		),
		sessionsClient: NewSessionsClient(apiAddress, apiToken, opts),
		usersClient:    NewUsersClient(apiAddress, apiToken, opts),
	}
}

func (a *apiClient) ServiceAccounts() ServiceAccountsClient {
	return a.serviceAccountsClient
}

func (a *apiClient) Sessions() SessionsClient {
	return a.sessionsClient
}

func (a *apiClient) Users() UsersClient {
	return a.usersClient
}
