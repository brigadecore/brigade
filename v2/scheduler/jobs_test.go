package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v3"
	coreTesting "github.com/brigadecore/brigade/sdk/v3/testing"
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
				substrateClient: &coreTesting.MockSubstrateClient{
					CountRunningJobsFn: func(
						context.Context,
						*sdk.RunningJobCountOptions,
					) (sdk.SubstrateJobCount, error) {
						return sdk.SubstrateJobCount{}, errors.New("something went wrong")
					},
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
				substrateClient: &coreTesting.MockSubstrateClient{
					CountRunningJobsFn: func(
						context.Context,
						*sdk.RunningJobCountOptions,
					) (sdk.SubstrateJobCount, error) {
						return sdk.SubstrateJobCount{
							Count: 2,
						}, nil
					},
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
				substrateClient: &coreTesting.MockSubstrateClient{
					CountRunningJobsFn: func(
						context.Context,
						*sdk.RunningJobCountOptions,
					) (sdk.SubstrateJobCount, error) {
						return sdk.SubstrateJobCount{
							Count: 1,
						}, nil
					},
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
						err, ok := i[0].(error)
						require.True(t, ok)
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
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(
							context.Context,
							string,
							*sdk.EventGetOptions,
						) (sdk.Event, error) {
							return sdk.Event{}, errors.New("something went wrong")
						},
					},
					jobsClient: &coreTesting.MockJobsClient{
						UpdateStatusFn: func(
							context.Context,
							string,
							string,
							sdk.JobStatus,
							*sdk.JobStatusUpdateOptions,
						) error {
							return nil
						},
					},
					jobLoopErrFn: func(i ...interface{}) {
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
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(
							context.Context,
							string,
							*sdk.EventGetOptions,
						) (sdk.Event, error) {
							return sdk.Event{
								Worker: &sdk.Worker{},
							}, nil
						},
					},
					jobLoopErrFn: func(i ...interface{}) {
						err, ok := i[0].(error)
						require.True(t, ok)
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
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(
							context.Context,
							string,
							*sdk.EventGetOptions,
						) (sdk.Event, error) {
							cancelFn()
							return sdk.Event{
								Worker: &sdk.Worker{
									Jobs: []sdk.Job{
										{
											Name: "bar",
											Status: &sdk.JobStatus{
												Phase: sdk.JobPhaseRunning,
											},
										},
									},
								},
							}, nil
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
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(
							context.Context,
							string,
							*sdk.EventGetOptions,
						) (sdk.Event, error) {
							return sdk.Event{
								Worker: &sdk.Worker{
									Jobs: []sdk.Job{
										{
											Name: "bar",
											Status: &sdk.JobStatus{
												Phase: sdk.JobPhasePending,
											},
										},
									},
								},
							}, nil
						},
					},
					jobAvailabilityCh: jobAvailabilityCh,
					jobsClient: &coreTesting.MockJobsClient{
						StartFn: func(
							context.Context,
							string,
							string,
							*sdk.JobStartOptions,
						) error {
							return errors.New("something went wrong")
						},
					},
					jobLoopErrFn: func(i ...interface{}) {
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
					eventsClient: &coreTesting.MockEventsClient{
						GetFn: func(
							context.Context,
							string,
							*sdk.EventGetOptions,
						) (sdk.Event, error) {
							return sdk.Event{
								Worker: &sdk.Worker{
									Jobs: []sdk.Job{
										{
											Name: "bar",
											Status: &sdk.JobStatus{
												Phase: sdk.JobPhasePending,
											},
										},
									},
								},
							}, nil
						},
					},
					jobAvailabilityCh: jobAvailabilityCh,
					jobsClient: &coreTesting.MockJobsClient{
						StartFn: func(
							context.Context,
							string,
							string,
							*sdk.JobStartOptions,
						) error {
							cancelFn()
							return nil
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
