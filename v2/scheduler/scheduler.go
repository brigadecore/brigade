package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/brigadecore/brigade-foundations/os"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
)

type schedulerConfig struct {
	healthcheckInterval          time.Duration
	addAndRemoveProjectsInterval time.Duration
	maxConcurrentWorkers         int
	maxConcurrentJobs            int
}

func getSchedulerConfig() (schedulerConfig, error) {
	config := schedulerConfig{}
	var err error
	config.healthcheckInterval = 30 * time.Second
	config.addAndRemoveProjectsInterval, err =
		os.GetDurationFromEnvVar("ADD_REMOVE_PROJECT_INTERVAL", 30*time.Second)
	if err != nil {
		return config, err
	}
	config.maxConcurrentWorkers, err =
		os.GetIntFromEnvVar("MAX_CONCURRENT_WORKERS", 1)
	if err != nil {
		return config, err
	}
	config.maxConcurrentJobs, err = os.GetIntFromEnvVar("MAX_CONCURRENT_JOBS", 3)
	return config, err
}

type scheduler struct {
	queueReaderFactory   queue.ReaderFactory
	projectsClient       core.ProjectsClient
	substrateClient      core.SubstrateClient
	eventsClient         core.EventsClient
	workersClient        core.WorkersClient
	jobsClient           core.JobsClient
	config               schedulerConfig
	workerAvailabilityCh chan struct{}
	jobAvailabilityCh    chan struct{}
	// All of the scheduler's goroutines will send fatal errors here
	errCh chan error
	// All of these internal functions are overridable for testing purposes
	manageWorkerCapacityFn func(context.Context)
	workerLoopErrFn        func(...interface{})
	manageJobCapacityFn    func(context.Context)
	jobLoopErrFn           func(...interface{})
	manageProjectsFn       func(context.Context)
	runHealthcheckLoopFn   func(ctx context.Context)
	runWorkerLoopFn        func(ctx context.Context, projectID string)
	runJobLoopFn           func(ctx context.Context, projectID string)
}

func newScheduler(
	coreClient core.APIClient,
	queueReaderFactory queue.ReaderFactory,
	config schedulerConfig,
) *scheduler {
	s := &scheduler{
		queueReaderFactory:   queueReaderFactory,
		config:               config,
		projectsClient:       coreClient.Projects(),
		substrateClient:      coreClient.Substrate(),
		eventsClient:         coreClient.Events(),
		workersClient:        coreClient.Events().Workers(),
		jobsClient:           coreClient.Events().Workers().Jobs(),
		workerAvailabilityCh: make(chan struct{}),
		jobAvailabilityCh:    make(chan struct{}),
		errCh:                make(chan error),
	}
	s.manageWorkerCapacityFn = s.manageWorkerCapacity
	s.workerLoopErrFn = log.Println
	s.manageJobCapacityFn = s.manageJobCapacity
	s.jobLoopErrFn = log.Println
	s.manageProjectsFn = s.manageProjects
	s.runHealthcheckLoopFn = s.runHealthcheckLoop
	s.runWorkerLoopFn = s.runWorkerLoop
	s.runJobLoopFn = s.runJobLoop
	return s
}

// run coordinates the many goroutines involved in different aspects of the
// scheduler. If any one of these goroutines encounters an unrecoverable error,
// everything shuts down.
func (s *scheduler) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	// Run healthcheck loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.runHealthcheckLoopFn(ctx)
	}()

	// Manage available Worker capacity
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.manageWorkerCapacityFn(ctx)
	}()

	// Manage available Job capacity
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.manageJobCapacityFn(ctx)
	}()

	// Monitor for new/deleted projects at a regular interval. Launch new or kill
	// existing project-specific Worker and Job loops as needed.
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.manageProjectsFn(ctx)
	}()

	// Wait for an error or a completed context
	var err error
	select {
	// In essence, this comprises the Scheduler's "healthcheck" logic.
	// Whenever we receive an error on this channel, we cancel the context and
	// shut down.  E.g., if one loop fails, everything fails.
	// This includes:
	//   1. an error listing projects at the start of the project loop
	//      (Scheduler <-> API comms)
	//   2. an error instantiating a reader in any of the worker/job loops
	//   3. an error instantiating a reader, reading a message or acking a
	//      message in the healhcheck loop, which runs regularly and may spot
	//      connectivity issues when the Brigade server is otherwise dormant
	//      (Scheduler <-> Messaging queue comms)
	case err = <-s.errCh:
		cancel() // Shut it all down
	case <-ctx.Done():
		err = ctx.Err()
	}

	// Adapt wg to a channel that can be used in a select
	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)
		wg.Wait()
	}()

	select {
	case <-doneCh:
	case <-time.After(3 * time.Second):
		// Probably doesn't matter that this is hardcoded. Relatively speaking, 3
		// seconds is a lot of time for things to wrap up.
	}

	return err
}
