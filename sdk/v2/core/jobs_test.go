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

func TestJobSpecMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, JobSpec{}, "JobSpec")
}

func TestJobStatusMarshalJSON(t *testing.T) {
	requireAPIVersionAndType(t, JobStatus{}, "JobStatus")
}

func TestJobsClientCreate(t *testing.T) {
	const testEventID = "12345"
	const testJobName = "Italian"
	testJob := Job{
		Spec: JobSpec{
			PrimaryContainer: JobContainerSpec{
				ContainerSpec: ContainerSpec{
					Image: "debian:latest",
				},
			},
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/events/%s/worker/jobs/%s",
						testEventID,
						testJobName,
					),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				job := Job{}
				err = json.Unmarshal(bodyBytes, &job)
				require.NoError(t, err)
				require.Equal(t, testJob, job)
				w.WriteHeader(http.StatusCreated)
			},
		),
	)
	defer server.Close()
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Create(
		context.Background(),
		testEventID,
		testJobName,
		testJob,
	)
	require.NoError(t, err)
}

func TestJobsClientStart(t *testing.T) {
	const testEventID = "12345"
	const testJobName = "Italian"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/events/%s/worker/jobs/%s/start",
						testEventID,
						testJobName,
					),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.Start(context.Background(), testEventID, testJobName)
	require.NoError(t, err)
}

func TestJobsClientGetStatus(t *testing.T) {
	const testEventID = "12345"
	const testJobName = "Italian"
	testJobStatus := JobStatus{
		Phase: JobPhaseRunning,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/events/%s/worker/jobs/%s/status",
						testEventID,
						testJobName,
					),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testJobStatus)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	jobStatus, err :=
		client.GetStatus(context.Background(), testEventID, testJobName)
	require.NoError(t, err)
	require.Equal(t, testJobStatus, jobStatus)
}

func TestJobsClientWatchStatus(t *testing.T) {
	const testEventID = "12345"
	const testJobName = "Italian"
	testStatus := JobStatus{
		Phase: JobPhaseRunning,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/events/%s/worker/jobs/%s/status",
						testEventID,
						testJobName,
					),
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
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	statusCh, _, err := client.WatchStatus(
		context.Background(),
		testEventID,
		testJobName,
	)
	require.NoError(t, err)
	select {
	case status := <-statusCh:
		require.Equal(t, testStatus, status)
	case <-time.After(3 * time.Second):
		require.Fail(t, "timed out waiting for status")
	}
}

func TestJobClientUpdateStatus(t *testing.T) {
	const testEventID = "12345"
	const testJobName = "Italian"
	testJobStatus := JobStatus{
		Phase: JobPhaseRunning,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf(
						"/v2/events/%s/worker/jobs/%s/status",
						testEventID,
						testJobName,
					),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				jobStatus := JobStatus{}
				err = json.Unmarshal(bodyBytes, &jobStatus)
				require.NoError(t, err)
				require.Equal(t, testJobStatus, jobStatus)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewJobsClient(
		server.URL,
		testAPIToken,
		testClientAllowInsecure,
	)
	err := client.UpdateStatus(
		context.Background(),
		testEventID,
		testJobName,
		testJobStatus,
	)
	require.NoError(t, err)
}
