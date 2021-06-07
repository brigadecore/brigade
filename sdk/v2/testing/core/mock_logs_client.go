package core

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/core"
)

type MockLogsClient struct {
	StreamFn func(
		ctx context.Context,
		eventID string,
		selector *core.LogsSelector,
		opts *core.LogStreamOptions,
	) (<-chan core.LogEntry, <-chan error, error)
}

func (m *MockLogsClient) Stream(
	ctx context.Context,
	eventID string,
	selector *core.LogsSelector,
	opts *core.LogStreamOptions,
) (<-chan core.LogEntry, <-chan error, error) {
	return m.StreamFn(ctx, eventID, selector, opts)
}
