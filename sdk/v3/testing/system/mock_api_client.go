package system

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3/system"
)

type MockAPIClient struct {
	PingFn func(
		context.Context,
		*system.PingOptions,
	) (system.PingResponse, error)
	UnversionedPingFn func(context.Context) ([]byte, error)
}

func (m *MockAPIClient) Ping(
	ctx context.Context,
	opts *system.PingOptions,
) (system.PingResponse, error) {
	return m.PingFn(ctx, opts)
}

func (m *MockAPIClient) UnversionedPing(ctx context.Context) ([]byte, error) {
	return m.UnversionedPingFn(ctx)
}
