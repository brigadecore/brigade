package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestGetObserverConfig(t *testing.T) {
	// Note that unit testing in Go does NOT clear environment variables between
	// tests, which can sometimes be a pain, but it's fine here-- so each of these
	// test cases builds on the previous case.
	testCases := []struct {
		name       string
		setup      func()
		assertions func(observerConfig, error)
	}{
		{
			name:  "success with defaults",
			setup: func() {},
			assertions: func(config observerConfig, err error) {
				require.Equal(t, time.Minute, config.delayBeforeCleanup)
			},
		},
		{
			name: "DELAY_BEFORE_CLEANUP not parsable as duration",
			setup: func() {
				os.Setenv("DELAY_BEFORE_CLEANUP", "foo")
			},
			assertions: func(config observerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a duration")
				require.Contains(t, err.Error(), "DELAY_BEFORE_CLEANUP")
			},
		},
		{
			name: "success with overrides",
			setup: func() {
				os.Setenv("DELAY_BEFORE_CLEANUP", "2m")
			},
			assertions: func(config observerConfig, err error) {
				require.Equal(t, 2*time.Minute, config.delayBeforeCleanup)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := getObserverConfig()
			testCase.assertions(config, err)
		})
	}
}

func TestNewObserver(t *testing.T) {
	workerClient := core.NewWorkersClient(
		"",
		"",
		&restmachinery.APIClientOptions{},
	)
	kubeClient := fake.NewSimpleClientset()
	config := observerConfig{
		delayBeforeCleanup: time.Minute,
	}
	observer := newObserver(workerClient, kubeClient, config)
	require.Same(t, kubeClient, observer.kubeClient)
	require.NotNil(t, observer.deletingPodsSet)
	require.NotNil(t, observer.syncMu)
	require.NotNil(t, observer.errCh)
	require.NotNil(t, observer.syncWorkerPodsFn)
	require.NotNil(t, observer.syncWorkerPodFn)
	require.NotNil(t, observer.syncJobPodsFn)
	require.NotNil(t, observer.syncJobPodFn)
	require.NotNil(t, observer.syncDeletedPodFn)
	require.NotNil(t, observer.updateWorkerStatusFn)
	require.NotNil(t, observer.cleanupWorkerFn)
	require.NotNil(t, observer.updateJobStatusFn)
	require.NotNil(t, observer.cleanupJobFn)
}

func TestObserverRun(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func() *observer
		assertions func(context.Context, error)
	}{
		{
			name: "worker pod sync produced error",
			setup: func() *observer {
				errCh := make(chan error)
				return &observer{
					syncWorkerPodsFn: func(context.Context) {
						errCh <- errors.New("something went wrong")
					},
					syncJobPodsFn: func(context.Context) {},
					errCh:         errCh,
				}
			},
			assertions: func(_ context.Context, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "job pod sync produced error",
			setup: func() *observer {
				errCh := make(chan error)
				return &observer{
					syncWorkerPodsFn: func(context.Context) {},
					syncJobPodsFn: func(context.Context) {
						errCh <- errors.New("something went wrong")
					},
					errCh: errCh,
				}
			},
			assertions: func(_ context.Context, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "context gets canceled",
			setup: func() *observer {
				return &observer{
					syncWorkerPodsFn: func(context.Context) {},
					syncJobPodsFn:    func(context.Context) {},
					errCh:            make(chan error),
				}
			},
			assertions: func(ctx context.Context, err error) {
				require.Error(t, err)
				require.Equal(t, ctx.Err(), err)
			},
		},
		{
			name: "timeout during shutdown",
			setup: func() *observer {
				return &observer{
					syncWorkerPodsFn: func(context.Context) {},
					syncJobPodsFn: func(context.Context) {
						// We'll make this function stubbornly never shut down. Everything
						// should still be ok.
						select {}
					},
					errCh: make(chan error),
				}
			},
			assertions: func(ctx context.Context, err error) {
				require.Error(t, err)
				require.Equal(t, ctx.Err(), err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err := testCase.setup().run(ctx)
			testCase.assertions(ctx, err)
		})
	}
}
