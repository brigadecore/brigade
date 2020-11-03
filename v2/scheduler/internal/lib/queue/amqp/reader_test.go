package amqp

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/Azure/go-amqp"
	myamqp "github.com/brigadecore/brigade/v2/internal/amqp"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
	"github.com/stretchr/testify/require"
)

func TestGetReaderFactoryConfig(t *testing.T) {
	// Note that unit testing in Go does NOT clear environment variables between
	// tests, which can sometimes be a pain, but it's fine here-- so each of these
	// test cases builds on the previous case.
	testCases := []struct {
		name       string
		setup      func()
		assertions func(ReaderFactoryConfig, error)
	}{
		{
			name:  "AMQP_ADDRESS not set",
			setup: func() {},
			assertions: func(_ ReaderFactoryConfig, err error) {
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
			assertions: func(_ ReaderFactoryConfig, err error) {
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
			assertions: func(_ ReaderFactoryConfig, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "value not found for")
				require.Contains(t, err.Error(), "AMQP_PASSWORD")
			},
		},
		{
			name: "success",
			setup: func() {
				os.Setenv("AMQP_PASSWORD", "bat")
			},
			assertions: func(config ReaderFactoryConfig, err error) {
				require.NoError(t, err)
				require.Equal(
					t,
					ReaderFactoryConfig{
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
			config, err := GetReaderFactoryConfig()
			testCase.assertions(config, err)
		})
	}
}

func TestNewReaderFactory(t *testing.T) {
	const testAddress = "foo"
	rf := NewReaderFactory(
		ReaderFactoryConfig{
			Address: testAddress,
		},
	)
	require.IsType(t, &readerFactory{}, rf)
	require.Equal(t, testAddress, rf.(*readerFactory).address)
	require.NotEmpty(t, rf.(*readerFactory).dialOpts)
	// Assert we're not connected yet. (It connects lazily.)
	require.Nil(t, rf.(*readerFactory).amqpClient)
	require.NotNil(t, rf.(*readerFactory).amqpClientMu)
	require.NotNil(t, rf.(*readerFactory).connectFn)
}

func TestReaderFactoryNewReader(t *testing.T) {
	const testQueueName = "foo"
	rf := &readerFactory{
		amqpClient: &mockAMQPClient{
			NewSessionFn: func(opts ...amqp.SessionOption) (myamqp.Session, error) {
				return &mockAMQPSession{
					NewReceiverFn: func(
						opts ...amqp.LinkOption,
					) (myamqp.Receiver, error) {
						return &mockAMQPReceiver{}, nil
					},
				}, nil
			},
		},
		amqpClientMu: &sync.Mutex{},
	}
	w, err := rf.NewReader(testQueueName)
	require.NoError(t, err)
	require.IsType(t, &reader{}, w)
	require.Equal(t, testQueueName, w.(*reader).queueName)
	require.NotNil(t, w.(*reader).amqpSession)
	require.NotNil(t, w.(*reader).amqpReceiver)
}

func TestReadFactoryClose(t *testing.T) {
	testCases := []struct {
		name          string
		readerFactory queue.ReaderFactory
		assertions    func(error)
	}{
		{
			name: "error closing underlying amqp connection",
			readerFactory: &readerFactory{
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
			readerFactory: &readerFactory{
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
			err := testCase.readerFactory.Close(context.Background())
			testCase.assertions(err)
		})
	}
}

func TestReaderRead(t *testing.T) {
	testCases := []struct {
		name       string
		reader     queue.Reader
		assertions func(error)
	}{
		{
			name: "error receiving message",
			reader: &reader{
				amqpReceiver: &mockAMQPReceiver{
					ReceiveFn: func(context.Context) (*amqp.Message, error) {
						return nil, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error receiving AMQP message")
			},
		},
		{
			name: "success",
			reader: &reader{
				amqpReceiver: &mockAMQPReceiver{
					ReceiveFn: func(context.Context) (*amqp.Message, error) {
						return &amqp.Message{}, nil
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
			_, err := testCase.reader.Read(context.Background())
			testCase.assertions(err)
		})
	}
}

func TestReaderClose(t *testing.T) {
	testCases := []struct {
		name       string
		reader     queue.Reader
		assertions func(error)
	}{
		{
			name: "error closing underlying receiver",
			reader: &reader{
				amqpReceiver: &mockAMQPReceiver{
					CloseFn: func(ctx context.Context) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error closing AMQP receiver")
			},
		},
		{
			name: "error closing underlying session",
			reader: &reader{
				amqpReceiver: &mockAMQPReceiver{
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
			reader: &reader{
				amqpReceiver: &mockAMQPReceiver{
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
			err := testCase.reader.Close(context.Background())
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

type mockAMQPReceiver struct {
	ReceiveFn func(ctx context.Context) (*amqp.Message, error)
	CloseFn   func(ctx context.Context) error
}

func (m *mockAMQPReceiver) Receive(ctx context.Context) (*amqp.Message, error) {
	return m.ReceiveFn(ctx)
}

func (m *mockAMQPReceiver) Close(ctx context.Context) error {
	return m.CloseFn(ctx)
}
