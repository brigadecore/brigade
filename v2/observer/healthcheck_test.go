package main

import (
	"context"
	"errors"
	"testing"
	"time"

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
				pingAPIServerFn: func(context.Context) error {
					return errors.New("something went wrong")
				},
			},
			assertions: func(err error) {
				require.Equal(t, "something went wrong", err.Error())
			},
		},

		{
			name: "success",
			observer: &observer{
				pingAPIServerFn: func(context.Context) error {
					return nil
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
