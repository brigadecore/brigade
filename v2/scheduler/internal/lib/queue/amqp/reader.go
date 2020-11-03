package amqp

import (
	"context"
	"sync"
	"time"

	"github.com/Azure/go-amqp"
	myamqp "github.com/brigadecore/brigade/v2/internal/amqp"
	"github.com/brigadecore/brigade/v2/internal/os"
	"github.com/brigadecore/brigade/v2/internal/retries"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
	"github.com/pkg/errors"
)

// ReaderFactoryConfig encapsulates details required for connecting an
// AMQP-based implementation of the queue.ReaderFactory interface to an
// underlying AMQP-based messaging service.
type ReaderFactoryConfig struct {
	// Address is the address of the AMQP-based messaging server.
	Address string
	// Username is the SASL username to use when connecting to the AMQP-based
	// messaging server.
	Username string
	// Password is the SASL password to use when connection to the AMQP-based
	// messaging server.
	Password string
}

// GetReaderFactoryConfig returns ReaderFactoryConfig based on configuration
// obtained from environment variables.
func GetReaderFactoryConfig() (ReaderFactoryConfig, error) {
	config := ReaderFactoryConfig{}
	var err error
	config.Address, err = os.GetRequiredEnvVar("AMQP_ADDRESS")
	if err != nil {
		return config, err
	}
	config.Username, err = os.GetRequiredEnvVar("AMQP_USERNAME")
	if err != nil {
		return config, err
	}
	config.Password, err = os.GetRequiredEnvVar("AMQP_PASSWORD")
	return config, err
}

// readerFactory is an AMQP-based implementation of the queue.ReaderFactory
// interface.
type readerFactory struct {
	address      string
	dialOpts     []amqp.ConnOption
	amqpClient   myamqp.Client
	amqpClientMu *sync.Mutex
	connectFn    func() error
}

// NewReaderFactory returns an an AMQP-based implementation of the
// queue.ReaderFactory interface.
func NewReaderFactory(config ReaderFactoryConfig) queue.ReaderFactory {
	q := &readerFactory{
		address: config.Address,
		dialOpts: []amqp.ConnOption{
			amqp.ConnSASLPlain(config.Username, config.Password),
		},
		amqpClientMu: &sync.Mutex{},
	}
	q.connectFn = q.connect
	return q
}

// connect connects (or reconnects) to the underlying AMQP-based messaging
// server. This function is NOT concurrency safe and callers should take
// measures to ensure they are the exclusive caller of this function.
func (q *readerFactory) connect() error {
	return retries.ManageRetries(
		context.Background(),
		"connect",
		10,
		10*time.Second,
		func() (bool, error) {
			if q.amqpClient != nil {
				q.amqpClient.Close()
			}
			var err error
			if q.amqpClient, err = myamqp.Dial(q.address, q.dialOpts...); err != nil {
				return true, errors.Wrap(err, "error dialing endpoint")
			}
			return false, nil
		},
	)
}

func (q *readerFactory) NewReader(queueName string) (queue.Reader, error) {
	// This entire function is a critical section of code so that we don't
	// possibly have multiple callers looking for a new Reader opening multiple
	// underlying connections to the messaging server.
	q.amqpClientMu.Lock()
	defer q.amqpClientMu.Unlock()

	if q.amqpClient == nil {
		if err := q.connectFn(); err != nil {
			return nil, err
		}
	}

	linkOpts := []amqp.LinkOption{
		amqp.LinkSourceAddress(queueName),
		// Link credit is 1 because we're a "slow" consumer. We do not want messages
		// piling up in a client-side buffer, knowing that it could be some time
		// before we can process them.
		amqp.LinkCredit(1),
	}

	// Every Writer will get its own Session and Receiver
	var amqpSession myamqp.Session
	var amqpReceiver myamqp.Receiver
	var err error
	for {
		// If we've been through the loop before, try cleaning up the session and/or
		// receiver that we never ended up using.
		if amqpReceiver != nil {
			amqpReceiver.Close(context.TODO()) // nolint: errcheck
		}
		if amqpSession != nil {
			amqpSession.Close(context.TODO()) // nolint: errcheck
		}

		if amqpSession, err = q.amqpClient.NewSession(); err != nil {
			// Assume this happened because the existing connection is no good. Try
			// to reconnect.
			if err = q.connectFn(); err != nil {
				// The connection function handles its own retries. If we got an error
				// here, it's pretty serious. Bail.
				return nil, err
			}
			// We're reconnected now, so loop around to try getting a session again.
			continue
		}
		if amqpReceiver, err = amqpSession.NewReceiver(linkOpts...); err != nil {
			// Assume this happened because the existing connection is no good.
			// Just loop around now because we not only need a new connection, but
			// also a new session.
			continue
		}
		break
	}

	return &reader{
		queueName:    queueName,
		amqpSession:  amqpSession,
		amqpReceiver: amqpReceiver,
	}, nil
}

func (q *readerFactory) Close(context.Context) error {
	if err := q.amqpClient.Close(); err != nil {
		return errors.Wrapf(err, "error closing AMQP client")
	}
	return nil
}

// reader is an AMQP-based implementation of the queue.Reader interface.
type reader struct {
	queueName    string
	amqpSession  myamqp.Session
	amqpReceiver myamqp.Receiver
}

func (q *reader) Read(
	ctx context.Context,
) (*queue.Message, error) {
	amqpMsg, err := q.amqpReceiver.Receive(ctx)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error receiving AMQP message for queue %q",
			q.queueName,
		)
	}
	return &queue.Message{
		Message: string(amqpMsg.GetData()),
		Ack:     amqpMsg.Accept,
	}, nil
}

func (q *reader) Close(ctx context.Context) error {
	if err := q.amqpReceiver.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP receiver for queue %q",
			q.queueName,
		)
	}
	if err := q.amqpSession.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP session for queue %q",
			q.queueName,
		)
	}
	return nil
}
