package main

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade-foundations/retries"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
)

// manageWorkerCapacity periodically checks how many Workers are currently
// running on the substrate and sends a signal on an availability channel when
// there is available capacity.
func (s *scheduler) manageWorkerCapacity(ctx context.Context) {
	for {
		// Look for capacity until we find some. Use a progressive backoff, capped
		// at 10 seconds between retries.
		if err := retries.ManageRetries(
			ctx,
			"find worker capacity",
			0,              // Infinite retries
			10*time.Second, // Max backoff
			func() (bool, error) {
				select {
				case <-ctx.Done():
					return false, nil // Stop looking
				default:
				}
				substrateWorkerCount, err := s.substrateClient.CountRunningWorkers(ctx)
				if err != nil {
					return false, err // Real error; stop retrying
				}
				if substrateWorkerCount.Count < s.config.maxConcurrentWorkers {
					return false, nil // Found capacity; stop looking
				}
				return true, nil // Keep looking
			},
		); err != nil {
			// Report error
			select {
			case s.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// Offer capacity to anyone who's ready for it
		select {
		case s.workerAvailabilityCh <- struct{}{}:
		case <-ctx.Done():
			return
		}

		// Wait to hear that whoever received the capacity has done whatever they
		// were going to do. This helps prevent a race where we loop around and
		// start looking for capacity and maybe find some that someone else is about
		// to claim.
		select {
		case <-s.workerAvailabilityCh:
		case <-ctx.Done():
			return
		}
	}
}

// nolint: gocyclo
func (s *scheduler) runWorkerLoop(ctx context.Context, projectID string) {

	var workersReader queue.Reader

outerLoop:
	for {

		if workersReader != nil {
			func() {
				closeCtx, cancelCloseCtx :=
					context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelCloseCtx()
				workersReader.Close(closeCtx)
			}()
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		workersReader, err = s.queueReaderFactory.NewReader(
			fmt.Sprintf("workers.%s", projectID),
		)
		if err != nil { // It's fatal if we can't get a queue reader
			select {
			case s.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// This is the main loop for receiving this Project's Events' Workers
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Get the next message
			msg, err := workersReader.Read(ctx)
			if err != nil {
				s.workerLoopErrFn(err)
				continue outerLoop // Try again with a new reader
			}

			eventID := msg.Message

			event, err := s.eventsClient.Get(ctx, eventID)
			if err != nil {
				s.workerLoopErrFn(err)

				if err := s.workersClient.UpdateStatus(
					ctx,
					eventID,
					core.WorkerStatus{
						Phase: core.WorkerPhaseSchedulingFailed,
					},
				); err != nil {
					s.workerLoopErrFn(err)
				}

				// Ack the message because there's nothing we can do and we don't want
				// something we can't process to clog the queue.
				if err := msg.Ack(ctx); err != nil {
					s.workerLoopErrFn(err)
				}
				continue // Next message
			}

			// If the Worker's phase isn't PENDING, then there's nothing to do
			if event.Worker.Status.Phase != core.WorkerPhasePending {
				if err := msg.Ack(ctx); err != nil {
					s.workerLoopErrFn(err)
				}
				continue // Next message
			}

			// Wait for capacity
			select {
			case <-s.workerAvailabilityCh:
			case <-ctx.Done():
				continue outerLoop // This will do cleanup before returning
			}

			// TODO: In the future, it would be nice to be able to block on PROJECT
			// capacity as well. i.e. Max of n concurrent workers per project. I
			// (krancour) see this as a post-2.0 enhancement.

			// Now use the API to start the Worker...

			if err := s.workersClient.Start(ctx, event.ID); err != nil {
				s.workerLoopErrFn(err)
			}

			if err := msg.Ack(ctx); err != nil {
				s.workerLoopErrFn(err)
			}

			// Tell the capacity manager we used the capacity it gave us
			select {
			case s.workerAvailabilityCh <- struct{}{}:
			case <-ctx.Done():
				continue outerLoop // This will do cleanup before returning
			}
		}

	}

}
