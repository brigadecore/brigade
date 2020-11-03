package queue

import "context"

// Message represents a message received from a queue (or similar channel) of
// some (presumably asynchronous) messaging system.
type Message struct {
	// Message is an unstructured, textual representation of the data.
	Message string
	// Ack is a function that may be invoked to (if applicable) signal the
	// underlying messaging system to consider the message delivered and
	// processed.
	Ack func(context.Context) error
}
