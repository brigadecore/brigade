package authn

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(rmTesting.TestAPIAddress, rmTesting.TestAPIToken, nil)
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
