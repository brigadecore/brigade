package system

import (
	"net/http"
	"testing"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"
)

// TODO: This isn't very DRY. It would be nice to figure out how to reuse these
// bits across a few different packages. The only way I (krancour) know of
// is to move these into their own package and NOT have them in files suffixed
// by _test.go. But were we to do that, Go would not recognize the functions as
// code used exclusively for testing and would therefore end up dinging us on
// coverage... for not testing the tests. :sigh:

const (
	testAPIAddress = "localhost:8080"
	testAPIToken   = "11235813213455"
)

func requireBaseClient(t *testing.T, baseClient *rm.BaseClient) {
	require.NotNil(t, baseClient)
	require.Equal(t, testAPIAddress, baseClient.APIAddress)
	require.Equal(t, testAPIToken, baseClient.APIToken)
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
