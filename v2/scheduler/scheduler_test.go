package main

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
	"github.com/stretchr/testify/require"
)

func TestGetSchedulerConfig(t *testing.T) {
	// Note that unit testing in Go does NOT clear environment variables between
	// tests, which can sometimes be a pain, but it's fine here-- so each of these
	// test cases builds on the previous case.
	testCases := []struct {
		name       string
		setup      func()
		assertions func(schedulerConfig, error)
	}{
		{
			name:  "success with defaults",
			setup: func() {},
			assertions: func(config schedulerConfig, err error) {
				require.Equal(t, 30*time.Second, config.addAndRemoveProjectsInterval)
				require.Equal(t, 1, config.maxConcurrentWorkers)
				require.Equal(t, 3, config.maxConcurrentJobs)
			},
		},
		{
			name: "ADD_REMOVE_PROJECT_INTERVAL not parsable as duration",
			setup: func() {
				os.Setenv("ADD_REMOVE_PROJECT_INTERVAL", "foo")
			},
			assertions: func(config schedulerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as a duration")
				require.Contains(t, err.Error(), "ADD_REMOVE_PROJECT_INTERVAL")
			},
		},
		{
			name: "MAX_CONCURRENT_WORKERS not parsable as int",
			setup: func() {
				os.Setenv("ADD_REMOVE_PROJECT_INTERVAL", "1m")
				os.Setenv("MAX_CONCURRENT_WORKERS", "foo")
			},
			assertions: func(config schedulerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as an int")
				require.Contains(t, err.Error(), "MAX_CONCURRENT_WORKERS")
			},
		},
		{
			name: "MAX_CONCURRENT_JOBS not parsable as int",
			setup: func() {
				os.Setenv("ADD_REMOVE_PROJECT_INTERVAL", "1m")
				os.Setenv("MAX_CONCURRENT_WORKERS", "5")
				os.Setenv("MAX_CONCURRENT_JOBS", "foo")
			},
			assertions: func(config schedulerConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "was not parsable as an int")
				require.Contains(t, err.Error(), "MAX_CONCURRENT_JOBS")
			},
		},
		{
			name: "success with overrides",
			setup: func() {
				os.Setenv("ADD_REMOVE_PROJECT_INTERVAL", "1m")
				os.Setenv("MAX_CONCURRENT_WORKERS", "5")
				os.Setenv("MAX_CONCURRENT_JOBS", "10")
			},
			assertions: func(config schedulerConfig, err error) {
				require.Equal(t, time.Minute, config.addAndRemoveProjectsInterval)
				require.Equal(t, 5, config.maxConcurrentWorkers)
				require.Equal(t, 10, config.maxConcurrentJobs)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := getSchedulerConfig()
			testCase.assertions(config, err)
		})
	}
}

func TestNewScheduler(t *testing.T) {
	coreClient := core.NewAPIClient("", "", &restmachinery.APIClientOptions{})
	queueReaderFactory := &mockQueueReaderFactory{}
	config := schedulerConfig{
		addAndRemoveProjectsInterval: 2 * time.Minute,
	}
	scheduler := newScheduler(coreClient, queueReaderFactory, config)
	require.Same(t, queueReaderFactory, scheduler.queueReaderFactory)
	require.NotNil(t, scheduler.projectsClient)
	require.NotNil(t, scheduler.substrateClient)
	require.NotNil(t, scheduler.eventsClient)
	require.NotNil(t, scheduler.workersClient)
	require.NotNil(t, scheduler.jobsClient)
	require.Equal(t, config, scheduler.config)
	require.NotNil(t, scheduler.workerAvailabilityCh)
	require.NotNil(t, scheduler.jobAvailabilityCh)
	require.NotNil(t, scheduler.errCh)
}

func TestSchedulerRun(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func() *scheduler
		assertions func(context.Context, error)
	}{
		{
			name: "healthcheck loop produced error",
			setup: func() *scheduler {
				s := &scheduler{
					manageJobCapacityFn:    func(context.Context) {},
					manageWorkerCapacityFn: func(context.Context) {},
					manageProjectsFn:       func(context.Context) {},
					errCh:                  make(chan error),
				}
				s.runHealthcheckLoopFn = func(context.Context) {
					s.errCh <- errors.New("something went wrong")
				}
				return s
			},
			assertions: func(_ context.Context, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "worker capacity manager produced error",
			setup: func() *scheduler {
				s := &scheduler{
					runHealthcheckLoopFn: func(context.Context) {},
					manageJobCapacityFn:  func(context.Context) {},
					manageProjectsFn:     func(context.Context) {},
					errCh:                make(chan error),
				}
				s.manageWorkerCapacityFn = func(context.Context) {
					s.errCh <- errors.New("something went wrong")
				}
				return s
			},
			assertions: func(_ context.Context, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "job capacity manager produced error",
			setup: func() *scheduler {
				s := &scheduler{
					runHealthcheckLoopFn:   func(context.Context) {},
					manageWorkerCapacityFn: func(context.Context) {},
					manageProjectsFn:       func(context.Context) {},
					errCh:                  make(chan error),
				}
				s.manageJobCapacityFn = func(context.Context) {
					s.errCh <- errors.New("something went wrong")
				}
				return s
			},
			assertions: func(_ context.Context, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "projects manager produced error",
			setup: func() *scheduler {
				s := &scheduler{
					runHealthcheckLoopFn:   func(context.Context) {},
					manageWorkerCapacityFn: func(context.Context) {},
					manageJobCapacityFn:    func(context.Context) {},
					errCh:                  make(chan error),
				}
				s.manageProjectsFn = func(context.Context) {
					s.errCh <- errors.New("something went wrong")
				}
				return s
			},
			assertions: func(_ context.Context, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
			},
		},
		{
			name: "context gets canceled",
			setup: func() *scheduler {
				return &scheduler{
					runHealthcheckLoopFn:   func(context.Context) {},
					manageWorkerCapacityFn: func(context.Context) {},
					manageJobCapacityFn:    func(context.Context) {},
					manageProjectsFn:       func(context.Context) {},
					errCh:                  make(chan error),
				}
			},
			assertions: func(ctx context.Context, err error) {
				require.Error(t, err)
				require.Equal(t, ctx.Err(), err)
			},
		},
		{
			name: "timeout during shutdown",
			setup: func() *scheduler {
				return &scheduler{
					runHealthcheckLoopFn:   func(context.Context) {},
					manageWorkerCapacityFn: func(context.Context) {},
					manageJobCapacityFn:    func(context.Context) {},
					manageProjectsFn: func(context.Context) {
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

type mockQueueReaderFactory struct {
	NewReaderFn func(queueName string) (queue.Reader, error)
	CloseFn     func(context.Context) error
}

func (m *mockQueueReaderFactory) NewReader(
	queueName string) (queue.Reader, error) {
	return m.NewReaderFn(queueName)
}

func (m *mockQueueReaderFactory) Close(ctx context.Context) error {
	return m.CloseFn(ctx)
}

type mockQueueReader struct {
	ReadFn  func(context.Context) (*queue.Message, error)
	CloseFn func(context.Context) error
}

func (m *mockQueueReader) Read(ctx context.Context) (*queue.Message, error) {
	return m.ReadFn(ctx)
}

func (m *mockQueueReader) Close(ctx context.Context) error {
	return m.CloseFn(ctx)
}
