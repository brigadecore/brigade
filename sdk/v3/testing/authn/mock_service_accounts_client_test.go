package authn

import (
	"testing"

	"github.com/brigadecore/brigade/sdk/v3/authn"
	"github.com/stretchr/testify/require"
)

func TestMockServiceAccountClient(t *testing.T) {
	require.Implements(
		t,
		(*authn.ServiceAccountsClient)(nil),
		&MockServiceAccountsClient{},
	)
}
