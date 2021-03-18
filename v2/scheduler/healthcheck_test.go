package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
	"github.com/stretchr/testify/require"
)

func TestRunHealthcheckLoop(t *testing.T) {
	testCases := []struct {
		name       string
		scheduler  *scheduler
		assertions func(error)
	}{
		{
			name: "error getting queue reader",
			scheduler: &scheduler{
				queueReaderFactory: &mockQueueReaderFactory{
					NewReaderFn: func(queueName string) (queue.Reader, error) {
						return nil, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Equal(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error reading from queue",
			scheduler: &scheduler{
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
			},
			assertions: func(err error) {
				require.Equal(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "error acking message",
			scheduler: &scheduler{
				queueReaderFactory: &mockQueueReaderFactory{
					NewReaderFn: func(queueName string) (queue.Reader, error) {
						return &mockQueueReader{
							ReadFn: func(c context.Context) (*queue.Message, error) {
								return &queue.Message{
									Message: "ping",
									Ack: func(context.Context) error {
										return errors.New("something went wrong")
									},
								}, nil
							},
							CloseFn: func(c context.Context) error {
								return nil
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Equal(t, err.Error(), "something went wrong")
			},
		},

		{
			name: "success",
			scheduler: &scheduler{
				queueReaderFactory: &mockQueueReaderFactory{
					NewReaderFn: func(queueName string) (queue.Reader, error) {
						return &mockQueueReader{
							ReadFn: func(c context.Context) (*queue.Message, error) {
								return &queue.Message{
									Message: "ping",
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
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			scheduler := testCase.scheduler
			scheduler.errCh = make(chan error)
			go scheduler.runHealthcheckLoop(ctx)
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
