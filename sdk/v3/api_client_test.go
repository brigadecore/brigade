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
	client := NewAPIClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &apiClient{}, client)
	require.NotNil(t, client.(*apiClient).authnClient)
	require.Equal(t, client.(*apiClient).authnClient, client.Authn())
	require.NotNil(t, client.(*apiClient).authzClient)
	require.Equal(t, client.(*apiClient).authzClient, client.Authz())
	require.NotNil(t, client.(*apiClient).coreClient)
	require.Equal(t, client.(*apiClient).coreClient, client.Core())
	require.NotNil(t, client.(*apiClient).systemClient)
	require.Equal(t, client.(*apiClient).systemClient, client.System())
}
