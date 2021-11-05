package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	coreTesting "github.com/brigadecore/brigade/sdk/v2/testing/core"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
	"github.com/stretchr/testify/require"
)

func TestManageWorkerCapacity(t *testing.T) {
	testCases := []struct {
		name       string
		scheduler  *scheduler
		assertions func(
			ctx context.Context,
			workerAvailabilityCh chan struct{},
			errCh chan error,
		)
	}{
		{
			name: "error checking capacity",
			scheduler: &scheduler{
				config: schedulerConfig{
					maxConcurrentWorkers: 2,
				},
				substrateClient: &coreTesting.MockSubstrateClient{
					CountRunningWorkersFn: func(
						context.Context,
					) (core.SubstrateWorkerCount, error) {
						return core.SubstrateWorkerCount{},
							errors.New("something went wrong")
					},
				},
				workerAvailabilityCh: make(chan struct{}),
				errCh:                make(chan error),
			},
			assertions: func(
				ctx context.Context,
				workerAvailabilityCh chan struct{},
				errCh chan error,
			) {
				select {
				case <-workerAvailabilityCh:
					require.Fail(
						t,
						"notified of available capacity when we should have received "+
							"an error",
					)
				case err := <-errCh:
					require.Error(t, err)
					require.Contains(t, err.Error(), "something went wrong")
				case <-ctx.Done():
					require.Fail(t, "never received expected error")
				}
			},
		},
		{
			name: "no capacity available",
			scheduler: &scheduler{
				config: schedulerConfig{
					maxConcurrentWorkers: 2,
				},
				substrateClient: &coreTesting.MockSubstrateClient{
					CountRunningWorkersFn: func(
						context.Context,
					) (core.SubstrateWorkerCount, error) {
						return core.SubstrateWorkerCount{
							Count: 2,
						}, nil
					},
				},
				workerAvailabilityCh: make(chan struct{}),
				errCh:                make(chan error),
			},
			assertions: func(
				ctx context.Context,
				workerAvailabilityCh chan struct{},
				errCh chan error,
			) {
				select {
				case <-workerAvailabilityCh:
					require.Fail(t, "notified of available capacity when none existed")
				case <-errCh:
					require.Fail(t, "received unexpected error")
				case <-ctx.Done():
				}
			},
		},
		{
			name: "capacity available",
			scheduler: &scheduler{
				config: schedulerConfig{
					maxConcurrentWorkers: 2,
				},
				substrateClient: &coreTesting.MockSubstrateClient{
					CountRunningWorkersFn: func(
						context.Context,
					) (core.SubstrateWorkerCount, error) {
						return core.SubstrateWorkerCount{
							Count: 1,
						}, nil
					},
				},
				workerAvailabilityCh: make(chan struct{}),
				errCh:                make(chan error),
			},
			assertions: func(
				ctx context.Context,
				workerAvailabilityCh chan struct{},
				errCh chan error,
			) {
				select {
				case <-workerAvailabilityCh:
					// Signal back to the capacity manager
					workerAvailabilityCh <- struct{}{}
				case <-errCh:
					require.Fail(t, "received unexpected error")
				case <-ctx.Done():
					require.Fail(t, "never notified of existing capacity")
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			go testCase.scheduler.manageWorkerCapacity(ctx)
			testCase.assertions(
				ctx,
				testCase.scheduler.workerAvailabilityCh,
				testCase.scheduler.errCh,
			)
			cancel()
		})
	}
}

func TestRunWorkerLoop(t *testing.T) {
	const testProject = "manhattan"
	testCases := []struct {
		name       string
		setup      func(ctx context.Context, cancelFn func()) *scheduler
		assertions func(error)
	}{
		{
			name: "error getting queue reader",
			setup: func(_ context.Context, cancelFn func()) *scheduler {
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return nil, errors.New("something went wrong")
						},
					},
					workerLoopErrFn: func(i ...interface{}) {
						require.Fail(
							t,
							"error logging function should not have been called",
						)
						cancelFn()
					},
				}
			},
			assertions: func(err error) {
				require.Equal(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error reading a message",
			setup: func(_ context.Context, cancelFn func()) *scheduler {
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return nil, errors.New("something went wrong")
								},
								CloseFn: func(c context.Context) error {
									return nil
								},
							}, nil
						},
					},
					workerLoopErrFn: func(i ...interface{}) {
						err, ok := i[0].(error)
						require.True(t, ok)
						require.Equal(t, err.Error(), "something went wrong")
						cancelFn()
					},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},

		{
			name: "error getting the event",
			setup: func(_ context.Context, cancelFn func()) *scheduler {
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Ack: func(context.Context) error {
											return nil
										},
									}, nil
								},
								CloseFn: func(c context.Context) error {
									return nil
								},
							}, nil
						},
					},
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(context.Context, string) (core.Event, error) {
							return core.Event{}, errors.New("something went wrong")
						},
					},
					workersClient: &coreTesting.MockWorkersClient{
						UpdateStatusFn: func(
							context.Context,
							string,
							core.WorkerStatus,
						) error {
							return nil
						},
					},
					workerLoopErrFn: func(i ...interface{}) {
						err, ok := i[0].(error)
						require.True(t, ok)
						require.Equal(t, err.Error(), "something went wrong")
						cancelFn()
					},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},

		{
			name: "worker phase is not PENDING",
			setup: func(_ context.Context, cancelFn func()) *scheduler {
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Ack: func(context.Context) error {
											return nil
										},
									}, nil
								},
								CloseFn: func(c context.Context) error {
									return nil
								},
							}, nil
						},
					},
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(context.Context, string) (core.Event, error) {
							cancelFn()
							return core.Event{
								Worker: &core.Worker{
									Status: core.WorkerStatus{
										Phase: core.WorkerPhaseRunning,
									},
								},
							}, nil
						},
					},
					workerLoopErrFn: func(i ...interface{}) {
						require.Fail(
							t,
							"error logging function should not have been called",
						)
						cancelFn()
					},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},

		{
			name: "error starting worker",
			setup: func(ctx context.Context, cancelFn func()) *scheduler {
				workerAvailabilityCh := make(chan struct{})
				go func() {
					select {
					case workerAvailabilityCh <- struct{}{}:
					case <-ctx.Done():
					}
					select {
					case <-workerAvailabilityCh:
					case <-ctx.Done():
					}
				}()
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Ack: func(context.Context) error {
											return nil
										},
									}, nil
								},
								CloseFn: func(c context.Context) error {
									return nil
								},
							}, nil
						},
					},
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(context.Context, string) (core.Event, error) {
							return core.Event{
								Worker: &core.Worker{
									Status: core.WorkerStatus{
										Phase: core.WorkerPhasePending,
									},
								},
							}, nil
						},
					},
					workerAvailabilityCh: workerAvailabilityCh,
					workersClient: &coreTesting.MockWorkersClient{
						StartFn: func(context.Context, string) error {
							return errors.New("something went wrong")
						},
					},
					workerLoopErrFn: func(i ...interface{}) {
						err, ok := i[0].(error)
						require.True(t, ok)
						require.Equal(t, err.Error(), "something went wrong")
						cancelFn()
					},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},

		{
			name: "success",
			setup: func(ctx context.Context, cancelFn func()) *scheduler {
				workerAvailabilityCh := make(chan struct{})
				go func() {
					select {
					case workerAvailabilityCh <- struct{}{}:
					case <-ctx.Done():
					}
					select {
					case <-workerAvailabilityCh:
					case <-ctx.Done():
					}
				}()
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Ack: func(context.Context) error {
											return nil
										},
									}, nil
								},
								CloseFn: func(c context.Context) error {
									return nil
								},
							}, nil
						},
					},
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(context.Context, string) (core.Event, error) {
							return core.Event{
								Worker: &core.Worker{
									Status: core.WorkerStatus{
										Phase: core.WorkerPhasePending,
									},
								},
							}, nil
						},
					},
					workerAvailabilityCh: workerAvailabilityCh,
					workersClient: &coreTesting.MockWorkersClient{
						StartFn: func(context.Context, string) error {
							cancelFn()
							return nil
						},
					},
					workerLoopErrFn: func(i ...interface{}) {
						require.Fail(
							t,
							"error logging function should not have been called",
						)
						cancelFn()
					},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			scheduler := testCase.setup(ctx, cancel)
			scheduler.errCh = make(chan error)
			go scheduler.runWorkerLoop(ctx, testProject)
			// Listen for errors
			select {
			case err := <-scheduler.errCh:
				testCase.assertions(err)
			case <-ctx.Done():
				testCase.assertions(nil)
			}
			cancel()
		})
	}
}
