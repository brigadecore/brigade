package testing

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/stretchr/testify/require"
)

func TestMockSubstrateClient(t *testing.T) {
	require.Implements(t, (*sdk.SubstrateClient)(nil), &MockSubstrateClient{})
}
