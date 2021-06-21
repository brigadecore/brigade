package core

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/stretchr/testify/require"
)

func TestMockWorkersClient(t *testing.T) {
	require.Implements(t, (*core.WorkersClient)(nil), &MockWorkersClient{})
}
