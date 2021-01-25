package core

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(rmTesting.TestAPIAddress, rmTesting.TestAPIToken, nil)
	require.IsType(t, &apiClient{}, client)
	require.NotNil(t, client.(*apiClient).projectsClient)
	require.Equal(t, client.(*apiClient).projectsClient, client.Projects())
	require.NotNil(t, client.(*apiClient).eventsClient)
	require.Equal(t, client.(*apiClient).eventsClient, client.Events())
	require.NotNil(t, client.(*apiClient).substrateClient)
	require.Equal(t, client.(*apiClient).substrateClient, client.Substrate())
}
