package main

import (
	"log"

	"github.com/brigadecore/brigade-foundations/signals"
	"github.com/brigadecore/brigade-foundations/version"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue/amqp"
)

func main() {
	log.Printf(
		"Starting Brigade Scheduler -- version %s -- commit %s",
		version.Version(),
		version.Commit(),
	)

	ctx := signals.Context()

	// Brigade core API client
	var apiClient core.APIClient
	{
		address, token, opts, err := apiClientConfig()
		if err != nil {
			log.Fatal(err)
		}
		apiClient = core.NewAPIClient(address, token, &opts)
	}

	// Message receiving abstraction
	var queueReaderFactory queue.ReaderFactory
	{
		config, err := readerFactoryConfig()
		if err != nil {
			log.Fatal(err)
		}
		queueReaderFactory = amqp.NewReaderFactory(config)
	}
	defer queueReaderFactory.Close(ctx)

	// Scheduler
	var scheduler *scheduler
	{
		config, err := getSchedulerConfig()
		if err != nil {
			log.Fatal(err)
		}
		scheduler = newScheduler(
			apiClient,
			queueReaderFactory,
			config,
		)
	}

	// Run it!
	log.Println(scheduler.run(ctx))
}
