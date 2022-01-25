package sdk

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewCoreClient(t *testing.T) {
	client := NewCoreClient(rmTesting.TestAPIAddress, rmTesting.TestAPIToken, nil)
	require.IsType(t, &coreClient{}, client)
	require.NotNil(t, client.(*coreClient).projectsClient)
	require.Equal(t, client.(*coreClient).projectsClient, client.Projects())
	require.NotNil(t, client.(*coreClient).eventsClient)
	require.Equal(t, client.(*coreClient).eventsClient, client.Events())
	require.NotNil(t, client.(*coreClient).substrateClient)
	require.Equal(t, client.(*coreClient).substrateClient, client.Substrate())
}
