package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery/testing" // nolint: lll
	"github.com/brigadecore/brigade/sdk/v2/meta"
	metaTesting "github.com/brigadecore/brigade/sdk/v2/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestEventMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, Event{}, EventLabel)
}

func TestEventListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, EventList{}, "EventList")
}

func TestNewEventsClient(t *testing.T) {
	client := NewEventsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	)
	require.IsType(t, &eventsClient{}, client)
	rmTesting.RequireBaseClient(t, client.(*eventsClient).BaseClient)
	require.NotNil(t, client.(*eventsClient).workersClient)
	require.Equal(t, client.(*eventsClient).workersClient, client.Workers())
	require.NotNil(t, client.(*eventsClient).logsClient)
	require.Equal(t, client.(*eventsClient).logsClient, client.Logs())
}

func TestEventsClientCreate(t *testing.T) {
	testEvent := Event{
		Payload: "a Tesla roadster",
	}
	testEvents := EventList{
		Items: []Event{
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "12345",
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "abcde",
				},
			},
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/events", r.URL.Path)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				event := Event{}
				err = json.Unmarshal(bodyBytes, &event)
				require.NoError(t, err)
				require.Equal(t, testEvent, event)
				bodyBytes, err = json.Marshal(testEvents)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	events, err := client.Create(
		context.Background(),
		testEvent,
	)
	require.NoError(t, err)
	require.Equal(t, testEvents, events)
}

func TestEventsClientList(t *testing.T) {
	const testProjectID = "bluebook"
	const testWorkerPhase = WorkerPhaseRunning
	testEvents := EventList{
		Items: []Event{
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "12345",
				},
			},
			{
				ObjectMeta: meta.ObjectMeta{
					ID: "abcde",
				},
			},
		},
	}

	t.Run("nil event selector", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, "/v2/events", r.URL.Path)
					require.Equal(t, 0, len(r.URL.Query()))
					bodyBytes, err := json.Marshal(testEvents)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, string(bodyBytes))
				},
			),
		)
		defer server.Close()
		client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
		events, err := client.List(
			context.Background(),
			nil,
			nil,
		)
		require.NoError(t, err)
		require.Equal(t, testEvents, events)
	})

	t.Run("non-nil event selector", func(t *testing.T) {
		server := httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					require.Equal(t, http.MethodGet, r.Method)
					require.Equal(t, "/v2/events", r.URL.Path)
					require.Equal(t, testProjectID, r.URL.Query().Get("projectID"))
					require.Equal(
						t,
						testWorkerPhase,
						WorkerPhase(r.URL.Query().Get("workerPhases")),
					)
					bodyBytes, err := json.Marshal(testEvents)
					require.NoError(t, err)
					w.WriteHeader(http.StatusOK)
					fmt.Fprintln(w, string(bodyBytes))
				},
			),
		)
		defer server.Close()
		client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
		events, err := client.List(
			context.Background(),
			&EventsSelector{
				ProjectID:    testProjectID,
				WorkerPhases: []WorkerPhase{WorkerPhaseRunning},
			},
			nil,
		)
		require.NoError(t, err)
		require.Equal(t, testEvents, events)
	})
}

func TestEventsClientGet(t *testing.T) {
	testEvent := Event{
		ObjectMeta: meta.ObjectMeta{
			ID: "12345",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s", testEvent.ID),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testEvent)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	event, err := client.Get(context.Background(), testEvent.ID)
	require.NoError(t, err)
	require.Equal(t, testEvent, event)
}

func TestEventsClientCancel(t *testing.T) {
	const testEventID = "12345"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/cancellation", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Cancel(context.Background(), testEventID)
	require.NoError(t, err)
}

func TestEventsClientCancelMany(t *testing.T) {
	const testProjectID = "bluebook"
	const testWorkerPhase = WorkerPhaseRunning
	testResult := CancelManyEventsResult{
		Count: 42,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(t, "/v2/events/cancellations", r.URL.Path)
				require.Equal(t, testProjectID, r.URL.Query().Get("projectID"))
				require.Equal(
					t,
					testWorkerPhase,
					WorkerPhase(r.URL.Query().Get("workerPhases")),
				)
				bodyBytes, err := json.Marshal(testResult)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	result, err := client.CancelMany(
		context.Background(),
		EventsSelector{
			ProjectID:    testProjectID,
			WorkerPhases: []WorkerPhase{WorkerPhaseRunning},
		},
	)
	require.NoError(t, err)
	require.Equal(t, testResult, result)
}

func TestEventsClientDelete(t *testing.T) {
	const testEventID = "12345"
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s", testEventID),
					r.URL.Path,
				)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.Delete(context.Background(), testEventID)
	require.NoError(t, err)
}

func TestEventsClientDeleteMany(t *testing.T) {
	const testProjectID = "bluebook"
	const testWorkerPhase = WorkerPhaseRunning
	testResult := DeleteManyEventsResult{
		Count: 42,
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, "/v2/events", r.URL.Path)
				require.Equal(t, testProjectID, r.URL.Query().Get("projectID"))
				require.Equal(
					t,
					testWorkerPhase,
					WorkerPhase(r.URL.Query().Get("workerPhases")),
				)
				bodyBytes, err := json.Marshal(testResult)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	result, err := client.DeleteMany(
		context.Background(),
		EventsSelector{
			ProjectID:    testProjectID,
			WorkerPhases: []WorkerPhase{WorkerPhaseRunning},
		},
	)
	require.NoError(t, err)
	require.Equal(t, testResult, result)
}
