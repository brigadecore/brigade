package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/v2/internal/os"
	"github.com/brigadecore/brigade/v2/scheduler/internal/lib/queue"
)

type schedulerConfig struct {
	addAndRemoveProjectsInterval time.Duration
	maxConcurrentWorkers         int
	maxConcurrentJobs            int
}

func getSchedulerConfig() (schedulerConfig, error) {
	config := schedulerConfig{}
	var err error
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
	runWorkerLoopFn        func(ctx context.Context, projectID string)
	runJobLoopFn           func(ctx context.Context, projectID string)
	// These normally point to API client functions, but can also be overridden
	// for test purposes
	listProjectsFn func(
		context.Context,
		*core.ProjectsSelector,
		*meta.ListOptions,
	) (core.ProjectList, error)
	countRunningWorkersFn func(context.Context) (core.SubstrateWorkerCount, error)
	countRunningJobsFn    func(context.Context) (core.SubstrateJobCount, error)
	getEventFn            func(context.Context, string) (core.Event, error)
	updateWorkerStatusFn  func(
		ctx context.Context,
		eventID string,
		status core.WorkerStatus,
	) error
	startWorkerFn     func(ctx context.Context, eventID string) error
	updateJobStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		status core.JobStatus,
	) error
	startJobFn func(ctx context.Context, eventID string, jobName string) error
}

func newScheduler(
	coreClient core.APIClient,
	queueReaderFactory queue.ReaderFactory,
	config schedulerConfig,
) *scheduler {
	s := &scheduler{
		queueReaderFactory:   queueReaderFactory,
		config:               config,
		workerAvailabilityCh: make(chan struct{}),
		jobAvailabilityCh:    make(chan struct{}),
		errCh:                make(chan error),
		// API functions
		listProjectsFn:        coreClient.Projects().List,
		countRunningWorkersFn: coreClient.Substrate().CountRunningWorkers,
		countRunningJobsFn:    coreClient.Substrate().CountRunningJobs,
		getEventFn:            coreClient.Events().Get,
		updateWorkerStatusFn:  coreClient.Events().Workers().UpdateStatus,
		startWorkerFn:         coreClient.Events().Workers().Start,
		updateJobStatusFn:     coreClient.Events().Workers().Jobs().UpdateStatus,
		startJobFn:            coreClient.Events().Workers().Jobs().Start,
	}
	s.manageWorkerCapacityFn = s.manageWorkerCapacity
	s.workerLoopErrFn = log.Println
	s.manageJobCapacityFn = s.manageJobCapacity
	s.jobLoopErrFn = log.Println
	s.manageProjectsFn = s.manageProjects
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
