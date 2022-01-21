package core

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/stretchr/testify/require"
)

func TestMockAuthzClient(t *testing.T) {
	require.Implements(t, (*core.AuthzClient)(nil), &MockAuthzClient{})
}
