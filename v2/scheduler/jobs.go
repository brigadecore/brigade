package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/brigadecore/brigade-foundations/retries"
	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
	"github.com/pkg/errors"
)

// manageJobCapacity periodically checks how many Jobs are currently running on
// the substrate and sends a signal on an availability channel when there is
// available capacity.
func (s *scheduler) manageJobCapacity(ctx context.Context) {
	for {
		// Look for capacity until we find some. Use a progressive backoff, capped
		// at 10 seconds between retries.
		if err := retries.ManageRetries(
			ctx,
			"find job capacity",
			0,              // Infinite retries
			10*time.Second, // Max backoff
			func() (bool, error) {
				select {
				case <-ctx.Done():
					return false, nil // Stop looking
				default:
				}
				substrateJobCount, err := s.substrateClient.CountRunningJobs(ctx, nil)
				if err != nil {
					return false, err // Real error; stop retrying
				}
				if substrateJobCount.Count < s.config.maxConcurrentJobs {
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
		case s.jobAvailabilityCh <- struct{}{}:
		case <-ctx.Done():
			return
		}

		// Wait to hear that whoever received the capacity has done whatever they
		// were going to do. This helps prevent a race where we loop around and
		// start looking for capacity and maybe find some that someone else is about
		// to claim.
		select {
		case <-s.jobAvailabilityCh:
		case <-ctx.Done():
			return
		}
	}
}

// nolint: gocyclo
func (s *scheduler) runJobLoop(ctx context.Context, projectID string) {

	var jobsReader queue.Reader

outerLoop:
	for {

		if jobsReader != nil {
			func() {
				closeCtx, cancelCloseCtx :=
					context.WithTimeout(context.Background(), 5*time.Second)
				defer cancelCloseCtx()
				jobsReader.Close(closeCtx)
			}()
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		var err error
		jobsReader, err = s.queueReaderFactory.NewReader(
			fmt.Sprintf("jobs.%s", projectID),
		)
		if err != nil { // It's fatal if we can't get a queue reader
			select {
			case s.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		// This is the main loop for receiving this Project's Events' Workers' Jobs
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Get the next message
			msg, err := jobsReader.Read(ctx)
			if err != nil {
				s.jobLoopErrFn(err)
				continue outerLoop // Try again with a new reader
			}

			messageTokens := strings.Split(msg.Message, ":")
			if len(messageTokens) != 2 {
				s.jobLoopErrFn(
					errors.Errorf(
						"received invalid message on project %q job queue",
						projectID,
					),
				)
				if err = msg.Ack(ctx); err != nil {
					s.jobLoopErrFn(err)
				}
				continue // Next message
			}
			eventID := messageTokens[0]
			jobName := messageTokens[1]

			event, err := s.eventsClient.Get(ctx, eventID, nil)
			if err != nil {
				s.jobLoopErrFn(err)

				if err := s.jobsClient.UpdateStatus(
					ctx,
					eventID,
					jobName,
					core.JobStatus{
						Phase: core.JobPhaseSchedulingFailed,
					},
					nil,
				); err != nil {
					s.jobLoopErrFn(err)
				}

				// Ack the message because there's nothing we can do and we don't want
				// something we can't process to clog the queue.
				if err := msg.Ack(ctx); err != nil {
					s.jobLoopErrFn(err)
				}
				continue // Next message
			}

			job, exists := event.Worker.Job(jobName)
			if !exists {
				s.jobLoopErrFn(
					errors.Errorf(
						"no job %q exists for event %q",
						jobName,
						eventID,
					),
				)
				if err := msg.Ack(ctx); err != nil {
					s.jobLoopErrFn(err)
				}
				continue // Next Job
			}

			// If the Job's phase isn't PENDING, then there's nothing to do
			if job.Status.Phase != core.JobPhasePending {
				if err := msg.Ack(ctx); err != nil {
					s.jobLoopErrFn(err)
				}
				continue // Next message
			}

			// Wait for capacity
			select {
			case <-s.jobAvailabilityCh:
			case <-ctx.Done():
				continue outerLoop // This will do cleanup before returning
			}

			// TODO: In the future, it would be nice to be able to block on PROJECT
			// capacity as well. i.e. Max of n concurrent jobs per project. I
			// (krancour) see this as a post-2.0 enhancement.

			// Now use the API to start the Job...

			if err := s.jobsClient.Start(ctx, event.ID, jobName, nil); err != nil {
				s.jobLoopErrFn(err)
			}

			if err := msg.Ack(ctx); err != nil {
				s.jobLoopErrFn(err)
			}

			// Tell the capacity manager we used the capacity it gave us
			select {
			case s.jobAvailabilityCh <- struct{}{}:
			case <-ctx.Done():
				continue outerLoop // This will do cleanup before returning
			}
		}

	}

}
