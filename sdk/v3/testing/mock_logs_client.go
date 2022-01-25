package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
)

type MockLogsClient struct {
	StreamFn func(
		ctx context.Context,
		eventID string,
		selector *sdk.LogsSelector,
		opts *sdk.LogStreamOptions,
	) (<-chan sdk.LogEntry, <-chan error, error)
}

func (m *MockLogsClient) Stream(
	ctx context.Context,
	eventID string,
	selector *sdk.LogsSelector,
	opts *sdk.LogStreamOptions,
) (<-chan sdk.LogEntry, <-chan error, error) {
	return m.StreamFn(ctx, eventID, selector, opts)
}
