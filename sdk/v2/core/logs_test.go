package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewLogsClient(t *testing.T) {
	client := NewLogsClient(testAPIAddress, testAPIToken, nil)
	require.IsType(t, &logsClient{}, client)
	requireBaseClient(t, client.(*logsClient).BaseClient)
}

func TestLogsClientStream(t *testing.T) {
	const testEventID = "12345"
	testSelector := LogsSelector{
		Job:       "farpoint",
		Container: "enterprise",
	}
	testOpts := LogStreamOptions{
		Follow: true,
	}
	testLogEntry := LogEntry{
		Message: "Captain's log, Stardate 41153.7. Our destination is Planet " +
			"Deneb IV, beyond which lies the great unexplored mass of the galaxy...",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodGet, r.Method)
				require.Equal(
					t,
					fmt.Sprintf("/v2/events/%s/logs", testEventID),
					r.URL.Path,
				)
				require.Equal(
					t,
					testSelector.Job,
					r.URL.Query().Get("job"),
				)
				require.Equal(
					t,
					testSelector.Container,
					r.URL.Query().Get("container"),
				)
				require.Equal(
					t,
					strconv.FormatBool(testOpts.Follow),
					r.URL.Query().Get("follow"),
				)
				bodyBytes, err := json.Marshal(testLogEntry)
				require.NoError(t, err)
				w.Header().Set("Content-Type", "text/event-stream")
				w.(http.Flusher).Flush()
				fmt.Fprintln(w, string(bodyBytes))
				w.(http.Flusher).Flush()
			},
		),
	)
	defer server.Close()
	client := NewLogsClient(server.URL, testAPIToken, nil)
	logsCh, _, err := client.Stream(
		context.Background(),
		testEventID,
		&testSelector,
		&testOpts,
	)
	require.NoError(t, err)
	select {
	case logEntry := <-logsCh:
		require.Equal(t, testLogEntry, logEntry)
	case <-time.After(3 * time.Second):
		require.Fail(t, "timed out waiting for logs")
	}
}
