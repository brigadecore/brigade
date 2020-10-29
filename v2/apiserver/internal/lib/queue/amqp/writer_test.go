package amqp

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/Azure/go-amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	myamqp "github.com/brigadecore/brigade/v2/internal/amqp"
	"github.com/stretchr/testify/require"
)

func TestGetWriterFactoryConfig(t *testing.T) {
	// Note that unit testing in Go does NOT clear environment variables between
	// tests, which can sometimes be a pain, but it's fine here-- so each of these
	// test cases builds on the previous case.
	testCases := []struct {
		name       string
		setup      func()
		assertions func(WriterFactoryConfig, error)
	}{
		{
			name:  "AMQP_ADDRESS not set",
			setup: func() {},
			assertions: func(_ WriterFactoryConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "AMQP_ADDRESS")
			},
		},
		{
			name: "AMQP_USERNAME not set",
			setup: func() {
				os.Setenv("AMQP_ADDRESS", "foo")
			},
			assertions: func(_ WriterFactoryConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "AMQP_USERNAME")
			},
		},
		{
			name: "AMQP_PASSWORD not set",
			setup: func() {
				os.Setenv("AMQP_USERNAME", "bar")
			},
			assertions: func(_ WriterFactoryConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "AMQP_PASSWORD")
			},
		},
		{
			name: "SUCCESS not set",
			setup: func() {
				os.Setenv("AMQP_PASSWORD", "bat")
			},
			assertions: func(config WriterFactoryConfig, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					WriterFactoryConfig{
						Address:  "foo",
						Username: "bar",
						Password: "bat",
					},
					config,
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup()
			config, err := GetWriterFactoryConfig()
			testCase.assertions(config, err)
		})
	}
}

func TestNewWriterFactory(t *testing.T) {
	const testAddress = "foo"
	wf := NewWriterFactory(
		WriterFactoryConfig{
			Address: testAddress,
		},
	)
	require.IsType(t, &writerFactory{}, wf)
	require.Equal(t, testAddress, wf.(*writerFactory).address)
	require.NotEmpty(t, wf.(*writerFactory).dialOpts)
	// Assert we're not connected yet. (It connects lazily.)
	require.Nil(t, wf.(*writerFactory).amqpClient)
	require.NotNil(t, wf.(*writerFactory).amqpClientMu)
	require.NotNil(t, wf.(*writerFactory).connectFn)
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
	require.IsType(t, &writer{}, w)
	require.Equal(t, testQueueName, w.(*writer).queueName)
	require.NotNil(t, w.(*writer).amqpSession)
	require.NotNil(t, w.(*writer).amqpSender)
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
					SendFn: func(context.Context, *amqp.Message) error {
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
			err := testCase.writer.Write(context.Background(), "message in a bottle")
			testCase.assertions(err)
		})
	}
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
