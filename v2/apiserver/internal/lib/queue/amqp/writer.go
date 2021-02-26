package amqp

import (
	"context"
	"sync"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	myamqp "github.com/brigadecore/brigade/v2/internal/amqp"
	"github.com/brigadecore/brigade/v2/internal/retries"
	"github.com/pkg/errors"
)

// WriterFactoryConfig encapsulates details required for connecting an
// AMQP-based implementation of the queue.WriterFactory interface to an
// underlying AMQP-based messaging service.
type WriterFactoryConfig struct {
	// Address is the address of the AMQP-based messaging server.
	Address string
	// Username is the SASL username to use when connecting to the AMQP-based
	// messaging server.
	Username string
	// Password is the SASL password to use when connection to the AMQP-based
	// messaging server.
	Password string
}

// writerFactory is an AMQP-based implementation of the queue.WriterFactory
// interface.
type writerFactory struct {
	address      string
	dialOpts     []amqp.ConnOption
	amqpClient   myamqp.Client
	amqpClientMu *sync.Mutex
	connectFn    func() error
}

// NewWriterFactory returns an an AMQP-based implementation of the
// queue.WriterFactory interface.
func NewWriterFactory(config WriterFactoryConfig) queue.WriterFactory {
	w := &writerFactory{
		address: config.Address,
		dialOpts: []amqp.ConnOption{
			amqp.ConnSASLPlain(config.Username, config.Password),
		},
		amqpClientMu: &sync.Mutex{},
	}
	w.connectFn = w.connect
	return w
}

// connect connects (or reconnects) to the underlying AMQP-based messaging
// server. This function is NOT concurrency safe and callers should take
// measures to ensure they are the exclusive caller of this function.
func (w *writerFactory) connect() error {
	return retries.ManageRetries(
		context.Background(),
		"connect",
		10,
		10*time.Second,
		func() (bool, error) {
			if w.amqpClient != nil {
				w.amqpClient.Close()
			}
			var err error
			if w.amqpClient, err = myamqp.Dial(w.address, w.dialOpts...); err != nil {
				return true, errors.Wrap(err, "error dialing endpoint")
			}
			return false, nil
		},
	)
}

func (w *writerFactory) NewWriter(queueName string) (queue.Writer, error) {
	// This entire function is a critical section of code so that we don't
	// possibly have multiple callers looking for a new Writer opening multiple
	// underlying connections to the messaging server.
	w.amqpClientMu.Lock()
	defer w.amqpClientMu.Unlock()

	if w.amqpClient == nil {
		if err := w.connectFn(); err != nil {
			return nil, err
		}
	}

	linkOpts := []amqp.LinkOption{
		amqp.LinkTargetAddress(queueName),
	}

	// Every Writer will get its own Session and Sender
	var amqpSession myamqp.Session
	var amqpSender myamqp.Sender
	var err error
	for {
		// If we've been through the loop before, try cleaning up the session and/or
		// sender that we never ended up using.
		if amqpSender != nil {
			amqpSender.Close(context.TODO()) // nolint: errcheck
		}
		if amqpSession != nil {
			amqpSession.Close(context.TODO()) // nolint: errcheck
		}

		if amqpSession, err = w.amqpClient.NewSession(); err != nil {
			// Assume this happened because the existing connection is no good. Try
			// to reconnect.
			if err = w.connectFn(); err != nil {
				// The connection function handles its own retries. If we got an error
				// here, it's pretty serious. Bail.
				return nil, err
			}
			// We're reconnected now, so loop around to try getting a session again.
			continue
		}
		if amqpSender, err = amqpSession.NewSender(linkOpts...); err != nil {
			// Assume this happened because the existing connection is no good.
			// Just loop around now because we not only need a new connection, but
			// also a new session.
			continue
		}
		break
	}

	return &writer{
		queueName:   queueName,
		amqpSession: amqpSession,
		amqpSender:  amqpSender,
	}, nil
}

func (w *writerFactory) Close(context.Context) error {
	if err := w.amqpClient.Close(); err != nil {
		return errors.Wrapf(err, "error closing AMQP client")
	}
	return nil
}

// writer is an AMQP-based implementation of the queue.Writer interface.
type writer struct {
	queueName   string
	amqpSession myamqp.Session
	amqpSender  myamqp.Sender
}

func (w *writer) Write(
	ctx context.Context,
	message string,
	opts *queue.MessageOptions,
) error {
	if opts == nil {
		opts = &queue.MessageOptions{Ephemeral: false}
	}
	msg := &amqp.Message{
		Header: &amqp.MessageHeader{
			Durable: !opts.Ephemeral,
		},
		Data: [][]byte{
			[]byte(message),
		},
	}
	if err := w.amqpSender.Send(ctx, msg); err != nil {
		return errors.Wrapf(
			err,
			"error sending amqp message for queue %q",
			w.queueName,
		)
	}
	return nil
}

func (w *writer) Close(ctx context.Context) error {
	if err := w.amqpSender.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP sender for queue %q",
			w.queueName,
		)
	}
	if err := w.amqpSession.Close(ctx); err != nil {
		return errors.Wrapf(
			err,
			"error closing AMQP session for queue %q",
			w.queueName,
		)
	}
	return nil
}
