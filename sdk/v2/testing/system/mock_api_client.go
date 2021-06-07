package system

import (
	"context"

	"github.com/brigadecore/brigade/sdk/v2/system"
)

type MockAPIClient struct {
	PingFn            func(ctx context.Context) (system.PingResponse, error)
	UnversionedPingFn func(ctx context.Context) ([]byte, error)
}

func (m *MockAPIClient) Ping(ctx context.Context) (system.PingResponse, error) {
	return m.PingFn(ctx)
}

func (m *MockAPIClient) UnversionedPing(ctx context.Context) ([]byte, error) {
	return m.UnversionedPingFn(ctx)
}
