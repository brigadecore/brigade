package main

import (
	"context"
	"time"
)

// runHealthcheckLoop checks Observer -> API Server comms
func (o *observer) runHealthcheckLoop(ctx context.Context) {
	ticker := time.NewTicker(o.config.healthcheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := o.pingAPIServerFn(ctx); err != nil {
				o.errCh <- err
			}
		default:
		}
	}
}
