package main

import (
	"context"

	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
)

// runHealthcheckLoop checks Scheduler -> Messaging queue comms
func (s *scheduler) runHealthcheckLoop(ctx context.Context) {

	var healthcheckReader queue.Reader
	var err error

	healthcheckReader, err = s.queueReaderFactory.NewReader("healthz")
	if err != nil { // It's fatal if we can't get a queue reader
		s.errCh <- err
	}

	// This is the main loop for receiving healthcheck pings
	for healthcheckReader != nil {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Get the next message on the queue
		msg, err := healthcheckReader.Read(ctx)
		if err != nil {
			// If we can't read from the queue, we consider it fatal
			// (Scheduler cancels the context when an error is received)
			s.errCh <- err
		}

		if msg != nil {
			if err = msg.Ack(ctx); err != nil {
				// If we can't ack a message, we consider it fatal
				// (Scheduler cancels the context when an error is received)
				s.errCh <- err
			}
		}
	}

}
