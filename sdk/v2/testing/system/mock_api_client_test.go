package system

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/system"
	"github.com/stretchr/testify/require"
)

func TestMockAPIClient(t *testing.T) {
	require.Implements(t, (*system.APIClient)(nil), &MockAPIClient{})
}
