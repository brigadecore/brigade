package amqp

import (
	"context"

	"github.com/Azure/go-amqp"
)

// Session is an interface for the subset of go-amqp *Session functions that we
// actually use, adapted slightly to also interact with our own custom Sender
// interface. Using these interfaces in our messaging abstraction, instead of
// using the go-amqp types directly, allows for the possibility of utilizing
// mock implementations for testing purposes. Adding only the subset of
// functions that we actually use limits the effort involved in creating such
// mocks.
type Session interface {
	// NewSender opens a new sender link on the session.
	NewSender(opts ...amqp.LinkOption) (Sender, error)
	// NewReceiver opens a new receiver link on the session.
	NewReceiver(opts ...amqp.LinkOption) (Receiver, error)
	// Close gracefully closes the session.
	Close(ctx context.Context) error
}

type session struct {
	session *amqp.Session
}

func (s *session) NewSender(opts ...amqp.LinkOption) (Sender, error) {
	sndr, err := s.session.NewSender(opts...)
	return &sender{
		sender: sndr,
	}, err
}

func (s *session) NewReceiver(opts ...amqp.LinkOption) (Receiver, error) {
	rcvr, err := s.session.NewReceiver(opts...)
	return &receiver{
		receiver: rcvr,
	}, err
}

func (s *session) Close(ctx context.Context) error {
	return s.session.Close(ctx)
}
