package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
	"github.com/stretchr/testify/require"
)

func TestManageJobsCapacity(t *testing.T) {
	testCases := []struct {
		name       string
		scheduler  *scheduler
		assertions func(
			ctx context.Context,
			jobAvailabilityCh chan struct{},
			errCh chan error,
		)
	}{
		{
			name: "error checking capacity",
			scheduler: &scheduler{
				config: schedulerConfig{
					maxConcurrentJobs: 2,
				},
				countRunningJobsFn: func(
					context.Context,
				) (core.SubstrateJobCount, error) {
					return core.SubstrateJobCount{}, errors.New("something went wrong")
				},
				jobAvailabilityCh: make(chan struct{}),
				errCh:             make(chan error),
			},
			assertions: func(
				ctx context.Context,
				jobAvailabilityCh chan struct{},
				errCh chan error,
			) {
				select {
				case <-jobAvailabilityCh:
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
					maxConcurrentJobs: 2,
				},
				countRunningJobsFn: func(
					context.Context,
				) (core.SubstrateJobCount, error) {
					return core.SubstrateJobCount{
						Count: 2,
					}, nil
				},
				jobAvailabilityCh: make(chan struct{}),
				errCh:             make(chan error),
			},
			assertions: func(
				ctx context.Context,
				jobAvailabilityCh chan struct{},
				errCh chan error,
			) {
				select {
				case <-jobAvailabilityCh:
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
					maxConcurrentJobs: 2,
				},
				countRunningJobsFn: func(
					context.Context,
				) (core.SubstrateJobCount, error) {
					return core.SubstrateJobCount{
						Count: 1,
					}, nil
				},
				jobAvailabilityCh: make(chan struct{}),
				errCh:             make(chan error),
			},
			assertions: func(
				ctx context.Context,
				jobAvailabilityCh chan struct{},
				errCh chan error,
			) {
				select {
				case <-jobAvailabilityCh:
					// Signal back to the capacity manager
					jobAvailabilityCh <- struct{}{}
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
			go testCase.scheduler.manageJobCapacity(ctx)
			testCase.assertions(
				ctx,
				testCase.scheduler.jobAvailabilityCh,
				testCase.scheduler.errCh,
			)
			cancel()
		})
	}
}

func TestRunJobLoop(t *testing.T) {
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
					jobLoopErrFn: func(i ...interface{}) {
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
					jobLoopErrFn: func(i ...interface{}) {
						err := i[0].(error)
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
			name: "message is invalid",
			setup: func(_ context.Context, cancelFn func()) *scheduler {
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Message: "foo", // Expected format is event:job
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
					jobLoopErrFn: func(i ...interface{}) {
						err := i[0].(error)
						require.Contains(t, err.Error(), "received invalid message")
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
										Message: "foo:bar",
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
					getEventFn: func(context.Context, string) (core.Event, error) {
						return core.Event{}, errors.New("something went wrong")
					},
					updateJobStatusFn: func(
						context.Context,
						string,
						string,
						core.JobStatus,
					) error {
						return nil
					},
					jobLoopErrFn: func(i ...interface{}) {
						err := i[0].(error)
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
			name: "job does not exist",
			setup: func(_ context.Context, cancelFn func()) *scheduler {
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Message: "foo:bar",
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
					getEventFn: func(context.Context, string) (core.Event, error) {
						return core.Event{
							Worker: &core.Worker{},
						}, nil
					},
					jobLoopErrFn: func(i ...interface{}) {
						err := i[0].(error)
						require.Contains(t, err.Error(), "no job")
						require.Contains(t, err.Error(), "exists for event")
						cancelFn()
					},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},

		{
			name: "job phase is not PENDING",
			setup: func(_ context.Context, cancelFn func()) *scheduler {
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Message: "foo:bar",
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
					getEventFn: func(context.Context, string) (core.Event, error) {
						cancelFn()
						return core.Event{
							Worker: &core.Worker{
								Jobs: []core.Job{
									{
										Name: "bar",
										Status: &core.JobStatus{
											Phase: core.JobPhaseRunning,
										},
									},
								},
							},
						}, nil
					},
					jobLoopErrFn: func(i ...interface{}) {
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
			name: "error starting job",
			setup: func(ctx context.Context, cancelFn func()) *scheduler {
				jobAvailabilityCh := make(chan struct{})
				go func() {
					select {
					case jobAvailabilityCh <- struct{}{}:
					case <-ctx.Done():
					}
					select {
					case <-jobAvailabilityCh:
					case <-ctx.Done():
					}
				}()
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Message: "foo:bar",
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
					getEventFn: func(context.Context, string) (core.Event, error) {
						return core.Event{
							Worker: &core.Worker{
								Jobs: []core.Job{
									{
										Name: "bar",
										Status: &core.JobStatus{
											Phase: core.JobPhasePending,
										},
									},
								},
							},
						}, nil
					},
					jobAvailabilityCh: jobAvailabilityCh,
					startJobFn: func(context.Context, string, string) error {
						return errors.New("something went wrong")
					},
					jobLoopErrFn: func(i ...interface{}) {
						err := i[0].(error)
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
				jobAvailabilityCh := make(chan struct{})
				go func() {
					select {
					case jobAvailabilityCh <- struct{}{}:
					case <-ctx.Done():
					}
					select {
					case <-jobAvailabilityCh:
					case <-ctx.Done():
					}
				}()
				return &scheduler{
					queueReaderFactory: &mockQueueReaderFactory{
						NewReaderFn: func(queueName string) (queue.Reader, error) {
							return &mockQueueReader{
								ReadFn: func(c context.Context) (*queue.Message, error) {
									return &queue.Message{
										Message: "foo:bar",
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
					getEventFn: func(context.Context, string) (core.Event, error) {
						return core.Event{
							Worker: &core.Worker{
								Jobs: []core.Job{
									{
										Name: "bar",
										Status: &core.JobStatus{
											Phase: core.JobPhasePending,
										},
									},
								},
							},
						}, nil
					},
					jobAvailabilityCh: jobAvailabilityCh,
					startJobFn: func(context.Context, string, string) error {
						cancelFn()
						return nil
					},
					jobLoopErrFn: func(i ...interface{}) {
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
			go scheduler.runJobLoop(ctx, testProject)
			// Listen for errors
			select {
			case err := <-scheduler.errCh:
				cancel()
				testCase.assertions(err)
			case <-ctx.Done():
			}
			cancel()
		})
	}
}
