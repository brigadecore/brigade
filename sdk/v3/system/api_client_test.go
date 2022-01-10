package system

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

func TestNewAPIClient(t *testing.T) {
	client := NewAPIClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &apiClient{}, client)
	rmTesting.RequireBaseClient(t, client.(*apiClient).BaseClient)
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
	client := NewAPIClient(server.URL, rmTesting.TestAPIToken, nil)
	resp, err := client.Ping(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, testResponse, resp)
}
