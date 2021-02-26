package main

import (
	"log"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/brigadecore/brigade/v2/internal/signals"
	"github.com/brigadecore/brigade/v2/internal/version"
)

// TODO: Observer needs functionality for timing out workers and jobs.

func main() {
	log.Printf(
		"Starting Brigade Observer -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	ctx := signals.Context()

	// Brigade Healthcheck and Workers API clients
	var healthcheckClient core.HealthcheckClient
	var workersClient core.WorkersClient
	{
		address, token, opts, err := apiClientConfig()
		if err != nil {
			log.Fatal(err)
		}
		healthcheckClient = core.NewHealthcheckClient(address, token, &opts)
		workersClient = core.NewWorkersClient(address, token, &opts)
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
			healthcheckClient,
			workersClient,
			kubeClient,
			config,
		)
	}

	// Run it!
	log.Println(observer.run(ctx))
}
