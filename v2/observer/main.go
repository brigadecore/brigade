package main

import (
	"context"
	"log"
	"time"

	"github.com/brigadecore/brigade-foundations/retries"
	"github.com/brigadecore/brigade-foundations/signals"
	"github.com/brigadecore/brigade-foundations/version"
	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/v2/internal/kubernetes"
)

func main() {
	log.Printf(
		"Starting Brigade Observer -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	ctx := signals.Context()

	// Brigade Healthcheck and Workers API clients
	var systemClient sdk.SystemClient
	var workersClient sdk.WorkersClient
	{
		address, token, opts, err := apiClientConfig()
		if err != nil {
			log.Fatal(err)
		}
		client := sdk.NewAPIClient(address, token, &opts)
		// No network I/O occurs when creating the API client, so we'll test it
		// here. This will block until the underlying connection is verified or max
		// retries are exhausted. What we're trying to prevent is both 1. moving on
		// in the startup process without the API server available AND 2. crashing
		// too prematurely while waiting for the API server to become available.
		if err := testClient(ctx, client); err != nil {
			log.Fatal(err)
		}
		systemClient = client.System()
		workersClient = client.Core().Events().Workers()
	}

	kubeClient, err := kubernetes.Client()
	if err != nil {
		log.Fatal(err)
	}

	// Observer
	var observer *observer
	{
		config, err := getObserverConfig()
		if err != nil {
			log.Fatal(err)
		}
		observer = newObserver(
			systemClient,
			workersClient,
			kubeClient,
			config,
		)
	}

	// Run it!
	log.Println(observer.run(ctx))
}

func testClient(ctx context.Context, client sdk.APIClient) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	return retries.ManageRetries(
		ctx, // The retry loop will exit when this context expires
		"api server ping",
		0,              // "Infinite" retries
		10*time.Second, // Max backoff
		func() (bool, error) {
			if _, err := client.System().Ping(ctx, nil); err != nil {
				return true, err // Retry
			}
			return false, nil // Success
		},
	)
}
