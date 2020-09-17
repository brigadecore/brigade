package authx

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/stretchr/testify/require"
)

// TODO: This isn't very DRY. It would be nice to figure out how to reuse these
// bits across a few different packages. The only way I (krancour) know of
// is to move these into their own package and NOT have them in files suffixed
// by _test.go. But were we to do that, Go would not recognize the functions as
// code used exclusively for testing and would therefore end up dinging us on
// coverage... for not testing the tests. :sigh:

const (
	testAPIAddress          = "localhost:8080"
	testAPIToken            = "11235813213455"
	testClientAllowInsecure = true
)

func requireAPIVersionAndType(
	t *testing.T,
	obj interface{},
	expectedType string,
) {
	objJSON, err := json.Marshal(obj)
	require.NoError(t, err)
	objMap := map[string]interface{}{}
	err = json.Unmarshal(objJSON, &objMap)
	require.NoError(t, err)
	require.Equal(t, meta.APIVersion, objMap["apiVersion"])
	require.Equal(t, expectedType, objMap["kind"])
}

func requireBaseClient(t *testing.T, baseClient *restmachinery.BaseClient) {
	require.Equal(t, testAPIAddress, baseClient.APIAddress)
	require.Equal(t, testAPIToken, baseClient.APIToken)
	require.IsType(t, &http.Client{}, baseClient.HTTPClient)
	require.IsType(t, &http.Transport{}, baseClient.HTTPClient.Transport)
	require.IsType(
		t,
		&tls.Config{},
		baseClient.HTTPClient.Transport.(*http.Transport).TLSClientConfig,
	)
	require.Equal(
		t,
		testClientAllowInsecure,
		baseClient.HTTPClient.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify, // nolint: lll
	)
}
