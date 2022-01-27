package testing

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/stretchr/testify/require"
)

func TestMockProjectAuthzClient(t *testing.T) {
	require.Implements(
		t,
		(*sdk.ProjectAuthzClient)(nil),
		&MockProjectAuthzClient{},
	)
}
