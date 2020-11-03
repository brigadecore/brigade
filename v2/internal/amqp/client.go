package amqp

import (
	"github.com/Azure/go-amqp"
)

// Client is an interface for the subset of go-amqp *Client functions that we
// actually use, adapted slightly to also interact with our own custom Session
// interface. Using these interfaces in our messaging abstraction, instead of
// using the go-amqp types directly, allows for the possibility of utilizing
// mock implementations for testing purposes. Adding only the subset of
// functions that we actually use limits the effort involved in creating such
// mocks.
type Client interface {
	// NewSession opens a new AMQP session to the server.
	NewSession(opts ...amqp.SessionOption) (Session, error)
	// Close disconnects the connection.
	Close() error
}

// client is an implementation of the Client interface that delegates all work
// to an underlying go-amqp *Client.
type client struct {
	client *amqp.Client
}

func (c *client) NewSession(opts ...amqp.SessionOption) (Session, error) {
	s, err := c.client.NewSession(opts...)
	return &session{
		session: s,
	}, err
}

func (c *client) Close() error {
	return c.client.Close()
}

// Dial connects to an AMQP server.
func Dial(addr string, opts ...amqp.ConnOption) (Client, error) {
	c, err := amqp.Dial(addr, opts...)
	return &client{
		client: c,
	}, err
}
