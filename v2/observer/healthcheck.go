package main

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// runHealthcheckLoop checks connectivity between the Observer and the two
// API servers it needs to communicate with: Brigade and K8s
func (o *observer) runHealthcheckLoop(ctx context.Context) {
	ticker := time.NewTicker(o.config.healthcheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check Observer -> API Server
			// Not actually capturing response; just want to verify our API call
			// is successful
			if _, err := o.systemClient.Ping(ctx); err != nil {
				o.errCh <- errors.Wrap(
					err,
					"error checking Brigade API server connectivity",
				)
			}

			// Check Observer -> K8s
			// Not actually capturing response; just want to verify our API call
			// is successful
			if _, err := o.checkK8sAPIServer(ctx); err != nil {
				o.errCh <- errors.Wrap(
					err,
					"error checking K8s API server connectivity",
				)
			}
		}
	}
}
