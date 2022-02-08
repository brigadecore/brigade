package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	metaTesting "github.com/brigadecore/brigade/sdk/v3/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestWorkerStatusMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, WorkerStatus{}, "WorkerStatus")
}

func TestNewWorkersClient(t *testing.T) {
	client, ok := NewWorkersClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*workersClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
	require.NotNil(t, client.jobsClient)
	require.Equal(t, client.jobsClient, client.Jobs())
}

func TestWorkersClientStart(t *testing.T) {
	const testEventID = "12345"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/start", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Start(context.Background(), testEventID, nil)
	require.NoError(t, err)
}

func TestWorkersClientGetStatus(t *testing.T) {
	const testEventID = "12345"
	testWorkerStatus := WorkerStatus{
		Phase: WorkerPhaseRunning,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/status", testEventID),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testWorkerStatus)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(server.URL, rmTesting.TestAPIToken, nil)
	workerStatus, err := client.GetStatus(context.Background(), testEventID, nil)
	require.NoError(t, err)
	require.Equal(t, testWorkerStatus, workerStatus)
}

func TestWorkersClientWatchStatus(t *testing.T) {
	const testEventID = "12345"
	testStatus := WorkerStatus{
		Phase: WorkerPhaseRunning,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/status", testEventID),
					r.URL.Path,
				)
				require.Equal(
					t,
					trueStr,
					r.URL.Query().Get("watch"),
				)
				bodyBytes, err := json.Marshal(testStatus)
				require.NoError(t, err)
				w.Header().Set("Content-Type", "text/event-stream")
				flusher, ok := w.(http.Flusher)
				require.True(t, ok)
				flusher.Flush()
				fmt.Fprintln(w, string(bodyBytes))
				flusher.Flush()
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(server.URL, rmTesting.TestAPIToken, nil)
	statusCh, _, err := client.WatchStatus(context.Background(), testEventID, nil)
	require.NoError(t, err)
	select {
	case status := <-statusCh:
		require.Equal(t, testStatus, status)
	case <-time.After(3 * time.Second):
		require.Fail(t, "timed out waiting for status")
	}
}

func TestWorkersClientUpdateStatus(t *testing.T) {
	const testEventID = "12345"
	testWorkerStatus := WorkerStatus{
		Phase: WorkerPhaseRunning,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/status", testEventID),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				workerStatus := WorkerStatus{}
				err = json.Unmarshal(bodyBytes, &workerStatus)
				require.NoError(t, err)
				require.Equal(t, testWorkerStatus, workerStatus)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.UpdateStatus(
		context.Background(),
		testEventID,
		testWorkerStatus,
		nil,
	)
	require.NoError(t, err)
}

func TestWorkersClientCleanup(t *testing.T) {
	const testEventID = "12345"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/cleanup", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Cleanup(context.Background(), testEventID, nil)
	require.NoError(t, err)
}

func TestWorkersClientTimeout(t *testing.T) {
	const testEventID = "12345"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/worker/timeout", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Timeout(context.Background(), testEventID, nil)
	require.NoError(t, err)
}
