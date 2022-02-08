package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/brigadecore/brigade/sdk/v3/meta"
	metaTesting "github.com/brigadecore/brigade/sdk/v3/meta/testing"
	"github.com/stretchr/testify/require"
)

func TestEventMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, Event{}, EventKind)
}

func TestEventListMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, EventList{}, "EventList")
}

func TestSourceStateMarshalJSON(t *testing.T) {
	metaTesting.RequireAPIVersionAndType(t, SourceState{}, "SourceState")
}

func TestNewEventsClient(t *testing.T) {
	client, ok := NewEventsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*eventsClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
	require.NotNil(t, client.workersClient)
	require.Equal(t, client.workersClient, client.Workers())
	require.NotNil(t, client.logsClient)
	require.Equal(t, client.logsClient, client.Logs())
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
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, testEvents, events)
}

func TestEventsClientList(t *testing.T) {
	const testProjectID = "bluebook"
	const testSource = "foo-gateway"
	const testType = "bar-type"
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
					require.Equal(t, testSource, r.URL.Query().Get("source"))
					require.Contains(t, r.URL.Query().Get("sourceState"), "foo=bar")
					require.Contains(t, r.URL.Query().Get("sourceState"), "bat=baz")
					require.Equal(t, testType, r.URL.Query().Get("type"))
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
				ProjectID: testProjectID,
				Source:    testSource,
				SourceState: map[string]string{
					"foo": "bar",
					"bat": "baz",
				},
				Type:         testType,
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
	event, err := client.Get(context.Background(), testEvent.ID, nil)
	require.NoError(t, err)
	require.Equal(t, testEvent, event)
}

func TestEventsClientClone(t *testing.T) {
	const testEventID = "12345"
	testEvent := Event{
		ObjectMeta: meta.ObjectMeta{
			ID: "12345",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/clones", testEventID),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testEvent)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	event, err := client.Clone(context.Background(), testEventID, nil)
	require.NoError(t, err)
	require.Equal(t, testEvent, event)
}

func TestEventsClientUpdateSourceState(t *testing.T) {
	const testEventID = "12345"
	testSourceState := SourceState{
		State: map[string]string{
			"foo": "bar",
			"bat": "baz",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/source-state", testEventID),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				sourceState := SourceState{}
				err = json.Unmarshal(bodyBytes, &sourceState)
				require.NoError(t, err)
				require.Equal(t, testSourceState, sourceState)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.UpdateSourceState(
		context.Background(),
		testEventID,
		testSourceState,
		nil,
	)
	require.NoError(t, err)
}

func TestEventsClientUpdateSummary(t *testing.T) {
	const testEventID = "12345"
	testSummary := EventSummary{
		Text: "foobar",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				require.Equal(t, http.MethodPut, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/summary", testEventID),
					r.URL.Path,
				)
				bodyBytes, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				summary := EventSummary{}
				err = json.Unmarshal(bodyBytes, &summary)
				require.NoError(t, err)
				require.Equal(t, testSummary, summary)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "{}")
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	err := client.UpdateSummary(
		context.Background(),
		testEventID,
		testSummary,
		nil,
	)
	require.NoError(t, err)
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
	err := client.Cancel(context.Background(), testEventID, nil)
	require.NoError(t, err)
}

func TestEventsClientCancelMany(t *testing.T) {
	const testProjectID = "bluebook"
	const testSource = "foo-gateway"
	const testType = "bar-event"
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
				require.Equal(t, testSource, r.URL.Query().Get("source"))
				require.Contains(t, r.URL.Query().Get("sourceState"), "foo=bar")
				require.Contains(t, r.URL.Query().Get("sourceState"), "bat=baz")
				require.Equal(t, testType, r.URL.Query().Get("type"))
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
			ProjectID: testProjectID,
			Source:    testSource,
			SourceState: map[string]string{
				"foo": "bar",
				"bat": "baz",
			},
			Type:         testType,
			WorkerPhases: []WorkerPhase{WorkerPhaseRunning},
		},
		nil,
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
	err := client.Delete(context.Background(), testEventID, nil)
	require.NoError(t, err)
}

func TestEventsClientDeleteMany(t *testing.T) {
	const testProjectID = "bluebook"
	const testSource = "foo-gateway"
	const testType = "bar-event"
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
				require.Equal(t, testSource, r.URL.Query().Get("source"))
				require.Contains(t, r.URL.Query().Get("sourceState"), "foo=bar")
				require.Contains(t, r.URL.Query().Get("sourceState"), "bat=baz")
				require.Equal(t, testType, r.URL.Query().Get("type"))
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
			ProjectID: testProjectID,
			Source:    testSource,
			SourceState: map[string]string{
				"foo": "bar",
				"bat": "baz",
			},
			Type:         testType,
			WorkerPhases: []WorkerPhase{WorkerPhaseRunning},
		},
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, testResult, result)
}

func TestEventsSelectorToQueryParams(t *testing.T) {
	testCases := []struct {
		name       string
		selector   *EventsSelector
		assertions func(map[string]string)
	}{
		{
			name: "nil selector",
			assertions: func(queryParams map[string]string) {
				require.Nil(t, queryParams)
			},
		},
		{
			name: "base case",
			selector: &EventsSelector{
				ProjectID: "blue-book",
				Source:    "brigade.sh/cli",
				Qualifiers: map[string]string{
					"foo": "bar",
					"bat": "baz",
				},
				Labels: map[string]string{
					"foo": "bar",
					"bat": "baz",
				},
				SourceState: map[string]string{
					"foo": "bar",
					"bat": "baz",
				},
				Type:         "exec",
				WorkerPhases: []WorkerPhase{WorkerPhasePending, WorkerPhaseStarting},
			},
			assertions: func(queryParams map[string]string) {
				qualifiers, ok := queryParams["qualifiers"]
				require.True(t, ok)
				require.Contains(t, qualifiers, "foo=bar")
				require.Contains(t, qualifiers, "bat=baz")
				labels, ok := queryParams["labels"]
				require.True(t, ok)
				require.Contains(t, labels, "foo=bar")
				require.Contains(t, labels, "bat=baz")
				sourceState, ok := queryParams["sourceState"]
				require.True(t, ok)
				require.Contains(t, sourceState, "foo=bar")
				require.Contains(t, sourceState, "bat=baz")
				delete(queryParams, "qualifiers")
				delete(queryParams, "labels")
				delete(queryParams, "sourceState")
				require.Equal(
					t,
					map[string]string{
						"projectID":    "blue-book",
						"source":       "brigade.sh/cli",
						"type":         "exec",
						"workerPhases": "PENDING,STARTING",
					},
					queryParams,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(eventsSelectorToQueryParams(testCase.selector))
		})
	}
}

func TestEventsClientRetry(t *testing.T) {
	const testEventID = "12345"
	testEvent := Event{
		ObjectMeta: meta.ObjectMeta{
			ID: "12345",
		},
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodPost, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/retries", testEventID),
					r.URL.Path,
				)
				bodyBytes, err := json.Marshal(testEvent)
				require.NoError(t, err)
				w.WriteHeader(http.StatusCreated)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	defer server.Close()
	client := NewEventsClient(server.URL, rmTesting.TestAPIToken, nil)
	event, err := client.Retry(context.Background(), testEventID, nil)
	require.NoError(t, err)
	require.Equal(t, testEvent, event)
}
