package amqp

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Azure/go-amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	myamqp "github.com/brigadecore/brigade/v2/internal/amqp"
	"github.com/stretchr/testify/require"
)

func TestNewWriterFactory(t *testing.T) {
	const testAddress = "foo"
	wf, ok := NewWriterFactory(
		WriterFactoryConfig{
			Address: testAddress,
		},
	).(*writerFactory)
	require.True(t, ok)
	require.Equal(t, testAddress, wf.address)
	require.NotEmpty(t, wf.dialOpts)
	// Assert we're not connected yet. (It connects lazily.)
	require.Nil(t, wf.amqpClient)
	require.NotNil(t, wf.amqpClientMu)
	require.NotNil(t, wf.connectFn)
}

func TestWriterFactoryNewWriter(t *testing.T) {
	const testQueueName = "foo"
	wf := &writerFactory{
		amqpClient: &mockAMQPClient{
			NewSessionFn: func(opts ...amqp.SessionOption) (myamqp.Session, error) {
				return &mockAMQPSession{
					NewSenderFn: func(opts ...amqp.LinkOption) (myamqp.Sender, error) {
						return &mockAMQPSender{}, nil
					},
				}, nil
			},
		},
		amqpClientMu: &sync.Mutex{},
	}
	w, err := wf.NewWriter(testQueueName)
	require.NoError(t, err)
	writer, ok := w.(*writer)
	require.True(t, ok)
	require.Equal(t, testQueueName, writer.queueName)
	require.NotNil(t, writer.amqpSession)
	require.NotNil(t, writer.amqpSender)
}

func TestWriterFactoryClose(t *testing.T) {
	testCases := []struct {
		name          string
		writerFactory queue.WriterFactory
		assertions    func(error)
	}{
		{
			name: "error closing underlying amqp connection",
			writerFactory: &writerFactory{
				amqpClient: &mockAMQPClient{
					CloseFn: func() error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error closing AMQP client")
			},
		},
		{
			name: "success",
			writerFactory: &writerFactory{
				amqpClient: &mockAMQPClient{
					CloseFn: func() error {
						return nil
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
			err := testCase.writerFactory.Close(context.Background())
			testCase.assertions(err)
		})
	}
}

func TestWriterWrite(t *testing.T) {
	testCases := []struct {
		name       string
		writer     queue.Writer
		assertions func(error)
	}{
		{
			name: "error sending message",
			writer: &writer{
				amqpSender: &mockAMQPSender{
					SendFn: func(context.Context, *amqp.Message) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error sending amqp message")
			},
		},
		{
			name: "success",
			writer: &writer{
				amqpSender: &mockAMQPSender{
					SendFn: func(ctx context.Context, msg *amqp.Message) error {
						if msg.Header.Durable != false {
							return errors.New("message persistence not as expected")
						}
						return nil
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
			err := testCase.writer.Write(
				context.Background(),
				"message in a bottle",
				&queue.MessageOptions{},
			)
			testCase.assertions(err)
		})
	}
}

func TestWriterWrite_MessageOpts(t *testing.T) {

	t.Run("nil opts", func(t *testing.T) {
		writer := &writer{
			amqpSender: &mockAMQPSender{
				SendFn: func(ctx context.Context, msg *amqp.Message) error {
					if msg.Header.Durable != false {
						return errors.New("message persistence not as expected")
					}
					return nil
				},
			},
		}
		err := writer.Write(
			context.Background(),
			"message in a bottle",
			nil,
		)
		require.NoError(t, err)
	})

	t.Run("non-nil opts", func(t *testing.T) {
		writer := &writer{
			amqpSender: &mockAMQPSender{
				SendFn: func(ctx context.Context, msg *amqp.Message) error {
					if msg.Header.Durable != true {
						return errors.New("message persistence not as expected")
					}
					return nil
				},
			},
		}
		err := writer.Write(
			context.Background(),
			"message in a bottle",
			&queue.MessageOptions{Durable: true},
		)
		require.NoError(t, err)
	})
}

func TestWriterClose(t *testing.T) {
	testCases := []struct {
		name       string
		writer     queue.Writer
		assertions func(error)
	}{
		{
			name: "error closing underlying sender",
			writer: &writer{
				amqpSender: &mockAMQPSender{
					CloseFn: func(ctx context.Context) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error closing AMQP sender")
			},
		},
		{
			name: "error closing underlying session",
			writer: &writer{
				amqpSender: &mockAMQPSender{
					CloseFn: func(ctx context.Context) error {
						return nil
					},
				},
				amqpSession: &mockAMQPSession{
					CloseFn: func(ctx context.Context) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error closing AMQP session")
			},
		},
		{
			name: "success",
			writer: &writer{
				amqpSender: &mockAMQPSender{
					CloseFn: func(ctx context.Context) error {
						return nil
					},
				},
				amqpSession: &mockAMQPSession{
					CloseFn: func(ctx context.Context) error {
						return nil
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
			err := testCase.writer.Close(context.Background())
			testCase.assertions(err)
		})
	}
}

type mockAMQPClient struct {
	NewSessionFn func(opts ...amqp.SessionOption) (myamqp.Session, error)
	CloseFn      func() error
}

func (m *mockAMQPClient) NewSession(
	opts ...amqp.SessionOption,
) (myamqp.Session, error) {
	return m.NewSessionFn(opts...)
}

func (m *mockAMQPClient) Close() error {
	return m.CloseFn()
}

type mockAMQPSession struct {
	NewSenderFn   func(opts ...amqp.LinkOption) (myamqp.Sender, error)
	NewReceiverFn func(opts ...amqp.LinkOption) (myamqp.Receiver, error)
	CloseFn       func(ctx context.Context) error
}

func (m *mockAMQPSession) NewSender(
	opts ...amqp.LinkOption,
) (myamqp.Sender, error) {
	return m.NewSenderFn(opts...)
}

func (m *mockAMQPSession) NewReceiver(
	opts ...amqp.LinkOption,
) (myamqp.Receiver, error) {
	return m.NewReceiverFn(opts...)
}

func (m *mockAMQPSession) Close(ctx context.Context) error {
	return m.CloseFn(ctx)
}

type mockAMQPSender struct {
	SendFn  func(ctx context.Context, msg *amqp.Message) error
	CloseFn func(ctx context.Context) error
}

func (m *mockAMQPSender) Send(ctx context.Context, msg *amqp.Message) error {
	return m.SendFn(ctx, msg)
}

func (m *mockAMQPSender) Close(ctx context.Context) error {
	return m.CloseFn(ctx)
}
