package main

import (
	"log"

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
		systemClient = sdk.NewSystemClient(address, token, &opts)
		workersClient = sdk.NewWorkersClient(address, token, &opts)
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
