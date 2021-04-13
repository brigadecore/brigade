package core

import (
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewAuthzClient(t *testing.T) {
	client := NewAuthzClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &authzClient{}, client)
	require.NotNil(t, client.(*authzClient).projectRoleAssignmentsClient)
	require.Equal(
		t,
		client.(*authzClient).projectRoleAssignmentsClient,
		client.RoleAssignments(),
	)
}
