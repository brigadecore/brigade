package amqp

import (
	"context"

	"github.com/Azure/go-amqp"
)

// Receiver is an interface for the subset of go-amqp *Receiver functions that
// we actually use. Using this interface in our messaging abstraction, instead
// of using the go-amqp type directly, allows for the possibility of utilizing
// mock implementations for testing purposes. Adding only the subset of
// functions that we actually use limits the effort involved in creating such
// mocks.
type Receiver interface {
	// Receive a Message.
	Receive(ctx context.Context) (*amqp.Message, error)
	// Close closes the Sender and AMQP link.
	Close(ctx context.Context) error
}
