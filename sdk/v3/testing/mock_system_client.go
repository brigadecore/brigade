package testing

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v3"
)

type MockSystemClient struct {
	PingFn func(
		context.Context,
		*sdk.PingOptions,
	) (sdk.PingResponse, error)
	UnversionedPingFn func(context.Context) ([]byte, error)
}

func (m *MockSystemClient) Ping(
	ctx context.Context,
	opts *sdk.PingOptions,
) (sdk.PingResponse, error) {
	return m.PingFn(ctx, opts)
}

func (m *MockSystemClient) UnversionedPing(
	ctx context.Context,
) ([]byte, error) {
	return m.UnversionedPingFn(ctx)
}
