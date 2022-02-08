package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewSystemClient(t *testing.T) {
	client, ok := NewSystemClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*systemClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
}

func TestAPIClientPing(t *testing.T) {
	testResponse := PingResponse{Version: "v2.0.0"}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					"/v2/ping",
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testResponse)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewSystemClient(server.URL, rmTesting.TestAPIToken, nil)
	resp, err := client.Ping(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, testResponse, resp)
}
