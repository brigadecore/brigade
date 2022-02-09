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

func TestNewSubstrateClient(t *testing.T) {
	client, ok := NewSubstrateClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*substrateClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
}

func TestSubstrateClientCountRunningWorkers(t *testing.T) {
	testCount := SubstrateWorkerCount{
		Count: 5,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/substrate/running-workers", r.URL.Path)
				bodyBytes, err := json.Marshal(testCount)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewSubstrateClient(server.URL, rmTesting.TestAPIToken, nil)
	count, err := client.CountRunningWorkers(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, testCount, count)
}

func TestSubstrateClientCountRunningJobs(t *testing.T) {
	testCount := SubstrateJobCount{
		Count: 5,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(t, "/v2/substrate/running-jobs", r.URL.Path)
				bodyBytes, err := json.Marshal(testCount)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewSubstrateClient(server.URL, rmTesting.TestAPIToken, nil)
	count, err := client.CountRunningJobs(context.Background(), nil)
	require.NoError(t, err)
	require.Equal(t, testCount, count)
}
