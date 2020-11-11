package retries

import (
	"context"
	"math"
	"time"

	"github.com/brigadecore/brigade/v2/internal/rand"
	"github.com/pkg/errors"
)

var seededRand = rand.NewSeeded()

var jitteredExpBackoffFn = jitteredExpBackoff

// ManageRetries executes the provided function until it "succeeds" or the
// maximum number of attempts has been exhausted. The delay between attempts
// increases exponentially, but is capped by the value of maxBackoff argument.
// "Success" requires the provided function to return a bool indicating that no
// retry should be attempted (false). The indication that a retry should or
// should not be attempted is SEPARATE from whatever error the provided function
// may return, with the decision to retry or not being made solely on the basis
// of the bool value. This makes it possible to retry failures after the
// provided function has done its own internal error handling. It also makes it
// possible to circumvent retries in the event that the error encountered is
// unrecoverable.
//
// Example usage:
//
// 	var client amqp.Client
// 	err := retries.ManageRetries(ctx, "amqp connect", 10, 10 * time.Second,
// 		func() (bool, error) {
// 			var e error
// 			if client, e = amqp.Dial(address); e != nil {
// 				return true, errors.Wrap(e, "error dialing AMQP server")
// 			}
// 			return false, nil
// 		},
// 	)
// 	// Handle err
func ManageRetries(
	ctx context.Context,
	process string,
	maxAttempts int,
	maxBackoff time.Duration,
	fn func() (bool, error),
) error {
	var failedAttempts int
	for {
		retry, err := fn()
		if !retry {
			return err
		}
		failedAttempts++
		if maxAttempts > 0 && failedAttempts == maxAttempts {
			if err == nil {
				return errors.Errorf(
					"failed %d attempt(s) to %s",
					maxAttempts,
					process,
				)
			}
			return errors.Wrapf(
				err,
				"failed %d attempt(s) to %s",
				maxAttempts,
				process,
			)
		}
		select {
		case <-time.After(jitteredExpBackoffFn(failedAttempts, maxBackoff)):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// jitteredExpBackoff returns a time.Duration to wait before the next retry when
// employing a "jittered" exponential backoff. The value returned is based,
// in-part on the number of failures to date and the maximum desired retry
// interval. This value is "jittered" before it is returned. The importance of
// this is that if many failures of any sort occur in rapid succession, the
// retries will be not only be staggered, but will become increasingly so as the
// failure count increases. This strategy helps to mitigate further
// complications in the event that the initial error was due to resource
// contention of rate limiting.
func jitteredExpBackoff(
	failureCount int,
	maxDelay time.Duration,
) time.Duration {
	base := math.Pow(2, float64(failureCount))
	capped := math.Min(base, maxDelay.Seconds())
	jittered := (1 + seededRand.Float64()) * (capped / 2)
	scaled := jittered * float64(time.Second)
	return time.Duration(scaled)
}
