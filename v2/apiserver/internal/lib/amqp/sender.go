package amqp

import (
	"context"

	"github.com/Azure/go-amqp"
)

// Sender is an interface for the subset of go-amqp *Sender functions that we
// actually use. Using this interface in our messaging abstraction, instead of
// using the go-amqp type directly, allows for the possibility of utilizing mock
// implementations for testing purposes. Adding only the subset of functions
// that we actually use limits the effort involved in creating such mocks.
type Sender interface {
	// Send sends a Message.
	Send(ctx context.Context, msg *amqp.Message) error
	// Close closes the Sender and AMQP link.
	Close(ctx context.Context) error
}

type sender struct {
	sender *amqp.Sender
}

func (s *sender) Send(ctx context.Context, msg *amqp.Message) error {
	return s.sender.Send(ctx, msg)
}

func (s *sender) Close(ctx context.Context) error {
	return s.sender.Close(ctx)
}
