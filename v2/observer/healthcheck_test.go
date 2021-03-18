package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/system"
	"github.com/stretchr/testify/require"
)

func TestRunHealthcheckLoop(t *testing.T) {
	testCases := []struct {
		name       string
		observer   *observer
		assertions func(error)
	}{
		{
			name: "error pinging API Server",
			observer: &observer{
				pingAPIServerFn: func(context.Context) (system.PingResponse, error) {
					return system.PingResponse{}, errors.New("something went wrong")
				},
				checkK8sAPIServer: func(context.Context) ([]byte, error) {
					return []byte{}, nil
				},
			},
			assertions: func(err error) {
				require.Contains(
					t,
					err.Error(),
					"error checking Brigade API server connectivity",
				)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error pinging K8s API Server",
			observer: &observer{
				pingAPIServerFn: func(context.Context) (system.PingResponse, error) {
					return system.PingResponse{}, nil
				},
				checkK8sAPIServer: func(context.Context) ([]byte, error) {
					return []byte{}, errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Contains(
					t,
					err.Error(),
					"error checking K8s API server connectivity",
				)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "success",
			observer: &observer{
				pingAPIServerFn: func(context.Context) (system.PingResponse, error) {
					return system.PingResponse{}, nil
				},
				checkK8sAPIServer: func(context.Context) ([]byte, error) {
					return []byte{}, nil
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			observer := testCase.observer
			observer.config = observerConfig{healthcheckInterval: time.Second}
			observer.errCh = make(chan error)
			go observer.runHealthcheckLoop(ctx)
			// Listen for errors
			select {
			case err := <-observer.errCh:
				cancel()
				testCase.assertions(err)
			case <-ctx.Done():
			}
			cancel()
		})
	}
}
