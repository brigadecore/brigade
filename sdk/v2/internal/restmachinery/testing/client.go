package testing

import (
	"net/http"
	"testing"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
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
	// The HTTPClient should be using a retriable transport...
	require.IsType(
		t,
		&retryablehttp.RoundTripper{},
		baseClient.HTTPClient.Transport,
	)
	// Which should delegate to a retriable client...
	require.NotNil(
		t,
		baseClient.HTTPClient.Transport.(*retryablehttp.RoundTripper).Client,
	)
	// Which uses a normal client...
	require.NotNil(
		t,
		baseClient.HTTPClient.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient, // nolint: lll
	)
	// With a normal transport...
	require.IsType(
		t,
		&http.Transport{},
		baseClient.HTTPClient.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient.Transport, // nolint: lll
	)
	// With TLS config
	require.NotNil(
		t,
		baseClient.HTTPClient.Transport.(*retryablehttp.RoundTripper).Client.HTTPClient.Transport.(*http.Transport).TLSClientConfig, // nolint: lll
	)
}
