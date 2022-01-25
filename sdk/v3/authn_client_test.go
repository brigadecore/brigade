package sdk

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewAuthnClient(t *testing.T) {
	client :=
		NewAuthnClient(rmTesting.TestAPIAddress, rmTesting.TestAPIToken, nil)
	require.IsType(t, &authnClient{}, client)
	require.NotNil(t, client.(*authnClient).serviceAccountsClient)
	require.Equal(
		t,
		client.(*authnClient).serviceAccountsClient,
		client.ServiceAccounts(),
	)
	require.NotNil(t, client.(*authnClient).sessionsClient)
	require.Equal(t, client.(*authnClient).sessionsClient, client.Sessions())
	require.NotNil(t, client.(*authnClient).usersClient)
	require.Equal(t, client.(*authnClient).usersClient, client.Users())
}
