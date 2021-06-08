package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/system"
	"github.com/brigadecore/brigade/v2/internal/os"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type observerConfig struct {
	delayBeforeCleanup  time.Duration
	healthcheckInterval time.Duration
}

func getObserverConfig() (observerConfig, error) {
	config := observerConfig{}
	var err error
	config.healthcheckInterval = 30 * time.Second
	config.delayBeforeCleanup, err =
		os.GetDurationFromEnvVar("DELAY_BEFORE_CLEANUP", time.Minute)
	return config, err
}

type observer struct {
	kubeClient      kubernetes.Interface
	systemClient    system.APIClient
	workersClient   core.WorkersClient
	jobsClient      core.JobsClient
	config          observerConfig
	deletingPodsSet map[string]struct{}
	timedPodsSet    map[string]context.CancelFunc
	syncMu          *sync.Mutex
	// All of the scheduler's goroutines will send fatal errors here
	errCh chan error
	// All of these internal functions are overridable for testing purposes
	runHealthcheckLoopFn    func(ctx context.Context)
	startJobPodTimerFn      func(ctx context.Context, pod *corev1.Pod)
	syncWorkerPodsFn        func(ctx context.Context)
	syncWorkerPodFn         func(obj interface{})
	deleteWorkerResourcesFn func(namespace, podName, eventID string)
	syncJobPodsFn           func(ctx context.Context)
	syncJobPodFn            func(obj interface{})
	deleteJobResourcesFn    func(namespace, podName, eventID, jobName string)
	syncDeletedPodFn        func(obj interface{})
	errFn                   func(...interface{})
	checkK8sAPIServer       func(ctx context.Context) ([]byte, error)
}

func newObserver(
	systemClient system.APIClient,
	workersClient core.WorkersClient,
	kubeClient kubernetes.Interface,
	config observerConfig,
) *observer {
	o := &observer{
		kubeClient:      kubeClient,
		systemClient:    systemClient,
		workersClient:   workersClient,
		jobsClient:      workersClient.Jobs(),
		config:          config,
		deletingPodsSet: map[string]struct{}{},
		timedPodsSet:    map[string]context.CancelFunc{},
		syncMu:          &sync.Mutex{},
		errCh:           make(chan error),
	}
	o.runHealthcheckLoopFn = o.runHealthcheckLoop
	o.startJobPodTimerFn = o.startJobPodTimer
	o.syncWorkerPodsFn = o.syncWorkerPods
	o.syncWorkerPodFn = o.syncWorkerPod
	o.deleteWorkerResourcesFn = o.deleteWorkerResources
	o.syncJobPodsFn = o.syncJobPods
	o.syncJobPodFn = o.syncJobPod
	o.deleteJobResourcesFn = o.deleteJobResources
	o.syncDeletedPodFn = o.syncDeletedPod
	o.errFn = log.Println

	// TODO: remove this type assertion once we figure out how to fake/mock
	// this k8s API Call
	if realK8sClient, ok := o.kubeClient.(*kubernetes.Clientset); ok {
		o.checkK8sAPIServer =
			realK8sClient.Discovery().RESTClient().Get().AbsPath("/healthz").DoRaw
	}
	return o
}

// run coordinates the many goroutines involved in different aspects of the
// observer. If any one of these goroutines encounters an unrecoverable error,
// everything shuts down.
func (o *observer) run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg := sync.WaitGroup{}

	// Run healthcheck loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.runHealthcheckLoopFn(ctx)
	}()

	// Continuously sync worker pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.syncWorkerPodsFn(ctx)
	}()

	// Continuously sync job pods
	wg.Add(1)
	go func() {
		defer wg.Done()
		o.syncJobPodsFn(ctx)
	}()

	// Wait for an error or a completed context
	var err error
	select {
	// In essence, this comprises the Observer's "healthcheck" logic.
	// Whenever we receive an error on this channel, we cancel the context and
	// shut down.  E.g., if one loop fails, everything fails.
	// This includes:
	//   1. an error pinging the Brigade API server endpoint
	//      (Observer <-> API comms)
	//   2. an error pinging the K8s API server endpoint
	//      (Observer <-> K8s comms)
	//
	// Note: Currently, errors updating or cleaning up worker or job statuses
	//       are handled by o.errFn, which currently simply logs the error
	case err = <-o.errCh:
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

// namespacedPodName is a utility function used by callers within this package
// to produce a map key from a given namespace name and pod name.
func namespacedPodName(namespace, name string) string {
	return fmt.Sprintf("%s:%s", namespace, name)
}
