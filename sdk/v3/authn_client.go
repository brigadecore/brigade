package sdk

import (
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
)

// AuthnClient is the root of a tree of more specialized API clients for dealing
// with identity and authentication.
type AuthnClient interface {
	// ServiceAccounts returns a specialized client for ServiceAccount management.
	ServiceAccounts() ServiceAccountsClient
	// Sessions returns a specialized client for Session management.
	Sessions() SessionsClient
	// Users returns a specialized client for User management.
	Users() UsersClient
}

type authnClient struct {
	// serviceAccountsClient is a specialized client for ServiceAccount
	// management.
	serviceAccountsClient ServiceAccountsClient
	// sessionsClient is a specialized client for Session management.
	sessionsClient SessionsClient
	// usersClient is a specialized client for User management.
	usersClient UsersClient
}

// NewAuthnClient returns an AuthnClient, which is the root of a tree of more
// specialized API clients for dealing with identity and authentication. It will
// initialize all clients in the tree so they are ready for immediate use.
func NewAuthnClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) AuthnClient {
	return &authnClient{
		serviceAccountsClient: NewServiceAccountsClient(
			apiAddress,
			apiToken,
			opts,
		),
		sessionsClient: NewSessionsClient(apiAddress, apiToken, opts),
		usersClient:    NewUsersClient(apiAddress, apiToken, opts),
	}
}

func (a *authnClient) ServiceAccounts() ServiceAccountsClient {
	return a.serviceAccountsClient
}

func (a *authnClient) Sessions() SessionsClient {
	return a.sessionsClient
}

func (a *authnClient) Users() UsersClient {
	return a.usersClient
}
