package queue

import "context"

// ReaderFactory is an interface for any component that can furnish an
// implementation of the Reader interface capable of reading messages from a
// specific queue (or similar channel) of some (presumably asynchronous)
// messaging system.
type ReaderFactory interface {
	// NewReader returns an implementation of the Reader interface capable of
	// reading messages from a specific queue (or similar channel) of some
	// (presumably asynchronous) messaging system.
	NewReader(queueName string) (Reader, error)
	// Close executes implementation-specific cleanup. Clients MUST invoke this
	// function when they are done with the ReaderFactory.
	Close(context.Context) error
}

// Reader is an interface used to abstract client code wishing to read messages
// from a specific queue (or similar channel) of some (presumably asynchronous)
// messaging system away from the underlying protocols and implementation of the
// mesaging system in use.
type Reader interface {
	// Read reads a message from a specific queue (or similar channel) known to
	// the implementation. This function blocks until either a successful read or
	// an error occurs.
	Read(ctx context.Context) (*Message, error)
	// Close executes implementation-specific cleanup. Clients MUST invoke this
	// function when they are done with the Reader.
	Close(context.Context) error
}
