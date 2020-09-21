package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWorkerStatusMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, WorkerStatus{}, "WorkerStatus")
}

func TestNewWorkersClient(t *testing.T) {
	client := NewWorkersClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &workersClient{}, client)
	requireBaseClient(t, client.(*workersClient).BaseClient)
	require.NotNil(t, client.(*workersClient).jobsClient)
	require.Equal(t, client.(*workersClient).jobsClient, client.Jobs())
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
	client := NewWorkersClient(server.URL, testAPIToken, nil)
	err := client.Start(context.Background(), testEventID)
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
	client := NewWorkersClient(server.URL, testAPIToken, nil)
	workerStatus, err := client.GetStatus(context.Background(), testEventID)
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
					"true",
					r.URL.Query().Get("watch"),
				)
				bodyBytes, err := json.Marshal(testStatus)
				require.NoError(t, err)
				w.Header().Set("Content-Type", "text/event-stream")
				w.(http.Flusher).Flush()
				fmt.Fprintln(w, string(bodyBytes))
				w.(http.Flusher).Flush()
			},
		),
	)
	defer server.Close()
	client := NewWorkersClient(server.URL, testAPIToken, nil)
	statusCh, _, err := client.WatchStatus(context.Background(), testEventID)
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
	client := NewWorkersClient(server.URL, testAPIToken, nil)
	err := client.UpdateStatus(
		context.Background(),
		testEventID,
		testWorkerStatus,
	)
	require.NoError(t, err)
}
