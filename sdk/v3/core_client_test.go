package sdk

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewCoreClient(t *testing.T) {
	client, ok := NewCoreClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*coreClient)
	require.True(t, ok)
	require.NotNil(t, client.projectsClient)
	require.Equal(t, client.projectsClient, client.Projects())
	require.NotNil(t, client.eventsClient)
	require.Equal(t, client.eventsClient, client.Events())
	require.NotNil(t, client.substrateClient)
	require.Equal(t, client.substrateClient, client.Substrate())
}
