package authx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(testAPIAddress, testAPIToken, testClientAllowInsecure)
	require.IsType(t, &apiClient{}, client)
	require.NotNil(t, client.(*apiClient).serviceAccountsClient)
	require.Equal(
		t,
		client.(*apiClient).serviceAccountsClient,
		client.ServiceAccounts(),
	)
	require.NotNil(t, client.(*apiClient).sessionsClient)
	require.Equal(t, client.(*apiClient).sessionsClient, client.Sessions())
	require.NotNil(t, client.(*apiClient).usersClient)
	require.Equal(t, client.(*apiClient).usersClient, client.Users())
}
