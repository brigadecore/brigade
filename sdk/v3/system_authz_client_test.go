package sdk

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewSystemAuthzClient(t *testing.T) {
	client :=
		NewSystemAuthzClient(rmTesting.TestAPIAddress, rmTesting.TestAPIToken, nil)
	require.NotNil(t, client.(*systemAuthzClient).roleAssignmentsClient)
	require.Equal(
		t,
		client.(*systemAuthzClient).roleAssignmentsClient,
		client.RoleAssignments(),
	)
}
