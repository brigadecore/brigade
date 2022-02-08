package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	rmTesting "github.com/brigadecore/brigade/sdk/v3/internal/restmachinery/testing" // nolint: lll
	"github.com/stretchr/testify/require"
)

func TestNewLogsClient(t *testing.T) {
	client, ok := NewLogsClient(
		rmTesting.TestAPIAddress,
		rmTesting.TestAPIToken,
		nil,
	).(*logsClient)
	require.True(t, ok)
	rmTesting.RequireBaseClient(t, client.BaseClient)
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

	t.Run("nil logs selector", func(t *testing.T) {
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
						1,
						len(r.URL.Query()),
					)
					require.Equal(
						t,
						strconv.FormatBool(testOpts.Follow),
						r.URL.Query().Get("follow"),
					)
					bodyBytes, err := json.Marshal(testLogEntry)
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
		client := NewLogsClient(server.URL, rmTesting.TestAPIToken, nil)
		logsCh, _, err := client.Stream(
			context.Background(),
			testEventID,
			nil,
			&testOpts,
		)
		require.NoError(t, err)
		select {
		case logEntry := <-logsCh:
			require.Equal(t, testLogEntry, logEntry)
		case <-time.After(3 * time.Second):
			require.Fail(t, "timed out waiting for logs")
		}
	})

	t.Run("non-nil logs selector", func(t *testing.T) {
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
					flusher, ok := w.(http.Flusher)
					require.True(t, ok)
					flusher.Flush()
					fmt.Fprintln(w, string(bodyBytes))
					flusher.Flush()
				},
			),
		)
		defer server.Close()
		client := NewLogsClient(server.URL, rmTesting.TestAPIToken, nil)
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
	})
}
