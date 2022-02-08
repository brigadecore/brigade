package testing

import (
	"net/http"
	"testing"

	rm "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

const (
	TestAPIAddress = "localhost:8080"
	TestAPIToken   = "11235813213455"
)

func RequireBaseClient(t *testing.T, baseClient *rm.BaseClient) {
	require.NotNil(t, baseClient)
	require.Equal(t, TestAPIAddress, baseClient.APIAddress)
	require.Equal(t, TestAPIToken, baseClient.APIToken)
	require.NotNil(t, baseClient.HTTPClient)
	roundTripper, ok :=
		baseClient.HTTPClient.Transport.(*retryablehttp.RoundTripper)
	require.True(t, ok)
	// Which should delegate to a retriable client...
	require.NotNil(t, roundTripper.Client)
	// Which uses a normal client...
	require.NotNil(t, roundTripper.Client.HTTPClient)
	// With a normal transport...
	require.IsType(t, &http.Transport{}, roundTripper.Client.HTTPClient.Transport)
	transport, ok := roundTripper.Client.HTTPClient.Transport.(*http.Transport)
	require.True(t, ok)
	// With TLS config
	require.NotNil(t, transport.TLSClientConfig)
}
