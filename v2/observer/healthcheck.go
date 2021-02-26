package main

import (
	"context"
	"time"

	"github.com/brigadecore/brigade/v2/internal/retries"
)

// runHealthcheckLoop checks Observer -> API Server comms
func (o *observer) runHealthcheckLoop(ctx context.Context) {
	ticker := time.NewTicker(o.config.healthcheckInterval)
	defer ticker.Stop()

	for {
		if err := retries.ManageRetries(
			ctx,
			"ping API server",
			5,
			10*time.Second, // Max backoff
			func() (bool, error) {
				select {
				case <-ctx.Done():
					return false, nil
				case <-ticker.C:
					if err := o.pingAPIServerFn(ctx); err != nil {
						return true, err
					}
					return false, nil
				}
			},
		); err != nil {
			// Send error along to the error channel, effectively shutting down
			select {
			case o.errCh <- err:
			case <-ctx.Done():
			}
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}
