package sdk

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewSystemAuthzClient(t *testing.T) {
	client, ok := NewSystemAuthzClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*systemAuthzClient)
	require.True(t, ok)
	require.Equal(t, client.roleAssignmentsClient, client.RoleAssignments())
}
