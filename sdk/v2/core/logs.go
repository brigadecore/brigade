package core

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/internal/restmachinery"
)

// LogsSelector represents useful criteria for selecting logs for streaming from
// a specific container of a Worker or Job.
type LogsSelector struct {
	// Job specifies, by name, a Job spawned by the Worker. If this field is
	// left blank, it is presumed logs are desired for the Worker itself.
	Job string
	// Container specifies, by name, a container belonging to the Worker or Job
	// whose logs are being retrieved. If left blank, a container with the same
	// name as the Worker or Job is assumed.
	Container string
}

// LogStreamOptions represents useful options for streaming logs from a specific
// container of a Worker or Job.
type LogStreamOptions struct {
	// Follow indicates whether the stream should conclude after the last
	// available line of logs has been sent to the client (false) or remain open
	// until closed by the client (true), continuing to send new lines as they
	// become available.
	Follow bool `json:"follow"`
}

// LogEntry represents one line of output from an OCI container.
type LogEntry struct {
	// Time is the time the line was written.
	Time *time.Time `json:"time,omitempty"`
	// Message is a single line of log output from an OCI container.
	Message string `json:"message,omitempty"`
}

// LogsClient is the specialized client for managing Logs with the Brigade API.
type LogsClient interface {
	// Stream returns a channel over which logs for an Event's Worker, or using
	// the LogsSelector parameter, a Job spawned by that Worker (or a specific
	// container thereof), are streamed.
	Stream(
		ctx context.Context,
		eventID string,
		selector LogsSelector,
		opts LogStreamOptions,
	) (<-chan LogEntry, <-chan error, error)
}

type logsClient struct {
	*restmachinery.BaseClient
}

// NewLogsClient returns a specialized client for managing Event Logs.
func NewLogsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) LogsClient {
	return &logsClient{
		BaseClient: &restmachinery.BaseClient{
			APIAddress: apiAddress,
			APIToken:   apiToken,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: allowInsecure,
					},
				},
			},
		},
	}
}

func (l *logsClient) Stream(
	ctx context.Context,
	eventID string,
	selector LogsSelector,
	opts LogStreamOptions,
) (<-chan LogEntry, <-chan error, error) {
	queryParams := map[string]string{}
	if selector.Job != "" {
		queryParams["job"] = selector.Job
	}
	if selector.Container != "" {
		queryParams["container"] = selector.Container
	}
	if opts.Follow {
		queryParams["follow"] = "true"
	}

	resp, err := l.SubmitRequest(
		ctx,
		restmachinery.OutboundRequest{
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
