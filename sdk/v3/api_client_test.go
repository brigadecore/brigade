package sdk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testAPIAddress = "localhost:8080"
	testAPIToken   = "11235813213455"
)

func TestNewAPIClient(t *testing.T) {
	client, ok := NewAPIClient(testAPIAddress, testAPIToken, nil).(*apiClient)
	require.True(t, ok)
	require.NotNil(t, client.authnClient)
	require.Equal(t, client.authnClient, client.Authn())
	require.NotNil(t, client.authzClient)
	require.Equal(t, client.authzClient, client.Authz())
	require.NotNil(t, client.coreClient)
	require.Equal(t, client.coreClient, client.Core())
	require.NotNil(t, client.systemClient)
	require.Equal(t, client.systemClient, client.System())
}
