package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	rm "github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

// LogEntry represents one line of output from an OCI container.
type LogEntry struct {
	// Time is the time the line was written.
	Time *time.Time `json:"time,omitempty"`
	// Message is a single line of log output from an OCI container.
	Message string `json:"message,omitempty"`
}

// LogsSelector represents useful criteria for selecting logs to be streamed
// from any container belonging to some Worker OR any container belonging to
// Jobs spawned by that Worker.
type LogsSelector struct {
	// Job specifies, by name, a Job spawned by some Worker. If not specified, log
	// streaming operations presume logs are desired for the Worker itself.
	Job string
	// Container specifies, by name, a container belonging to some Worker or, if
	// Job is specified, that Job. If not specified, log streaming operations
	// presume logs are desired from a container having the same name as the
	// selected Worker or Job.
	Container string
}

// LogStreamOptions represents useful options for streaming logs from some
// container of a Worker or Job.
type LogStreamOptions struct {
	// Follow indicates whether the stream should conclude after the last
	// available line of logs has been sent to the client (false) or remain open
	// until closed by the client (true), continuing to send new lines as they
	// become available.
	Follow bool `json:"follow"`
}

// LogsClient is the specialized client for managing Logs with the Brigade API.
type LogsClient interface {
	// Stream returns a channel over which logs for an Event's Worker, or using
	// the LogsSelector parameter, a Job spawned by that Worker (or a specific
	// container thereof), are streamed.
	Stream(
		ctx context.Context,
		eventID string,
		selector *LogsSelector,
		opts *LogStreamOptions,
	) (<-chan LogEntry, <-chan error, error)
}

type logsClient struct {
	*rm.BaseClient
}

// NewLogsClient returns a specialized client for managing Event Logs.
func NewLogsClient(
	apiAddress string,
	apiToken string,
	opts *restmachinery.APIClientOptions,
) LogsClient {
	return &logsClient{
		BaseClient: rm.NewBaseClient(apiAddress, apiToken, opts),
	}
}

func (l *logsClient) Stream(
	ctx context.Context,
	eventID string,
	selector *LogsSelector,
	opts *LogStreamOptions,
) (<-chan LogEntry, <-chan error, error) {
	queryParams := map[string]string{}
	if selector != nil {
		if selector.Job != "" {
			queryParams["job"] = selector.Job
		}
		if selector.Container != "" {
			queryParams["container"] = selector.Container
		}
	}
	if opts != nil && opts.Follow {
		queryParams["follow"] = "true"
	}

	resp, err := l.SubmitRequest(
		ctx,
		rm.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/events/%s/logs", eventID),
			AuthHeaders: l.BearerTokenAuthHeaders(),
			QueryParams: queryParams,
			SuccessCode: http.StatusOK,
		},
	)
	if err != nil {
		return nil, nil, err
	}

	logCh := make(chan LogEntry)
	errCh := make(chan error)

	go l.receiveStream(ctx, resp.Body, logCh, errCh)

	return logCh, errCh, nil
}

// receiveStream is used to receive log messages as SSEs (server sent events),
// decode those, and publish them to a channel.
func (l *logsClient) receiveStream(
	ctx context.Context,
	reader io.ReadCloser,
	logEntryCh chan<- LogEntry,
	errCh chan<- error,
) {
	defer close(logEntryCh)
	defer close(errCh)
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		logEntry := LogEntry{}
		if err := decoder.Decode(&logEntry); err != nil {
			if err == io.EOF {
				return
			}
			select {
			case errCh <- err:
			case <-ctx.Done():
			}
			return
		}
		select {
		case logEntryCh <- logEntry:
		case <-ctx.Done():
			return
		}
	}
}
