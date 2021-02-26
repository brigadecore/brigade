package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewHealthcheckClient(t *testing.T) {
	client := NewHealthcheckClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &healthcheckClient{}, client)
	rmTesting.RequireBaseClient(t, client.(*healthcheckClient).BaseClient)
}

func TestHealthcheckClientPing(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					"/ping",
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewHealthcheckClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Ping(context.Background())
	require.NoError(t, err)
}
