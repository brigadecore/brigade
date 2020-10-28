package queue

import "context"

// WriterFactory is an interface for any component that can furnish an
// implementation of the Writer interface capable of writing messages to a
// specific queue (or similar channel) of some (presumably asynchronous)
// messaging system.
type WriterFactory interface {
	// NewWriter returns an implementation of the Writer interface for capable of
	// writing messages to a specific queue (or similar channel) of some
	// (presumably asynchronous) messaging system.
	NewWriter(queueName string) (Writer, error)
	// Close executes implementation-specific cleanup. Clients MUST invoke this
	// function when they are done with the WriterFactory.
	Close(context.Context) error
}

// Writer is an interface used to abstract client code wishing to write messages
// to a specific queue (or similar channel) of some (presumably asynchronous)
// messaging system away from the underlying protocols and implementation of the
// mesaging system in use.
type Writer interface {
	// Write writes the provided message to a specific queue (or similar channel)
	// known to the implementation.
	Write(ctx context.Context, message string) error
	// Close executes implementation-specific cleanup. Clients MUST invoke this
	// function when they are done with the Writer.
	Close(context.Context) error
}
