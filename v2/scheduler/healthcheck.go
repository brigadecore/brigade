package main

import (
	"context"
	"time"

	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
)

// runHealthcheckLoop checks Scheduler -> Messaging queue comms
func (s *scheduler) runHealthcheckLoop(ctx context.Context) {

	var healthcheckReader queue.Reader

outerLoop:
	for {

		if healthcheckReader != nil {
			func() {
				closeCtx, cancelCloseCtx :=
					context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelCloseCtx()
				healthcheckReader.Close(closeCtx)
			}()
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		healthcheckReader, err = s.queueReaderFactory.NewReader("healthz")
		if err != nil { // It's fatal if we can't get a queue reader
			select {
			case s.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// This is the main loop for receiving healthcheck pings
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// TODO: do we want to allow for a certain threshold of retries here
			// before sending an error (and hence shutting down the Scheduler)?

			// Get the next message on the queue
			msg, err := healthcheckReader.Read(ctx)
			if err != nil {
				// If we can't read from the queue, we consider it fatal
				// (Scheduler cancels the context when an error is received)
				s.errCh <- err
				continue outerLoop // Try again with a new reader
			}

			if err = msg.Ack(ctx); err != nil {
				// If we can't ack a message, we consider it fatal
				// (Scheduler cancels the context when an error is received)
				s.errCh <- err
			}
			continue // Next message
		}

	}

}
