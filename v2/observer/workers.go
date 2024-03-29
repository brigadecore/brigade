package main

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (o *observer) syncWorkerPods(ctx context.Context) {
	workersSelector := myk8s.WorkerPodsSelector(o.config.brigadeID)
	workerPodsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = workersSelector
				return o.kubeClient.CoreV1().Pods("").List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = workersSelector
				return o.kubeClient.CoreV1().Pods("").Watch(ctx, options)
			},
		},
		&corev1.Pod{},
		0,
		cache.Indexers{},
	)
	workerPodsInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: o.syncWorkerPodFn,
			UpdateFunc: func(_, newObj interface{}) {
				o.syncWorkerPodFn(newObj)
			},
		},
	)
	workerPodsInformer.Run(ctx.Done())
}

func (o *observer) syncWorkerPod(obj interface{}) {
	ctx := context.Background()
	pod := obj.(*corev1.Pod) // nolint: forcetypeassert
	// Map pod status to worker status
	status := o.getWorkerStatusFromPod(pod)
	// Manage the timeout clock
	o.manageWorkerTimeoutFn(ctx, pod, status.Phase)
	// Use the API to update Worker status
	eventID := pod.Labels[myk8s.LabelEvent]
	ctx, cancel := context.WithTimeout(ctx, apiRequestTimeout)
	defer cancel()
	if err := o.workersClient.UpdateStatus(
		ctx,
		eventID,
		status,
		nil,
	); err != nil {
		if _, conflict :=
			err.(*meta.ErrConflict); !conflict || pod.DeletionTimestamp == nil {
			// Only log the error if it's NOT a conflict, or it is, but we're not
			// processing a delete. We expect conflicts that we can safely ignore
			// occur frequently for status updates on a delete.
			o.errFn(
				fmt.Sprintf(
					"error updating status for event %q worker: %s",
					eventID,
					err,
				),
			)
		}
	}
	if pod.DeletionTimestamp != nil {
		// If the pod was deleted, immediately complete cleanup of other resources
		// associated with the worker
		o.cleanupWorkerFn(eventID)
	} else if status.Phase.IsTerminal() {
		// Otherwise, if the worker is in a terminal phase, defer cleanup so the log
		// agent has a chance to catch up if necessary.
		go func() {
			<-time.After(o.config.delayBeforeCleanup)
			o.cleanupWorkerFn(eventID)
		}()
	}
}

func (o *observer) getWorkerStatusFromPod(pod *corev1.Pod) sdk.WorkerStatus {
	// Determine the worker's phase based on pod phase
	status := sdk.WorkerStatus{
		Phase: sdk.WorkerPhaseRunning,
	}
	if pod.DeletionTimestamp != nil {
		// This pod has been deleted. Pods usually hang around for a while after
		// they complete before they are deleted (to ensure the log agent has enough
		// time to capture all logs), so the worker's final transition to a terminal
		// phase has PROBABLY been recorded already, BUT given the possibility that
		// the worker's pod was manually deleted by an operator, we'll attempt a
		// final phase change to ABORTED here, knowing that it will simply fail if
		// the worker has already reached a terminal state.
		status.Phase = sdk.WorkerPhaseAborted
	} else {
		switch pod.Status.Phase {
		case corev1.PodPending:
			// For Brigade's purposes, this counts as running
			status.Phase = sdk.WorkerPhaseRunning
			// Unless... when an image pull backoff occurs, the pod still shows as
			// pending. We account for that here and treat it as a failure.
			//
			// TODO: Are there other conditions we need to watch out for?
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil &&
					(containerStatus.State.Waiting.Reason == "ImagePullBackOff" ||
						containerStatus.State.Waiting.Reason == "ErrImagePull") {
					status.Phase = sdk.WorkerPhaseFailed
					break
				}
			}
		case corev1.PodRunning:
			status.Phase = sdk.WorkerPhaseRunning
		case corev1.PodSucceeded:
			status.Phase = sdk.WorkerPhaseSucceeded
		case corev1.PodFailed:
			status.Phase = sdk.WorkerPhaseFailed
		case corev1.PodUnknown:
			status.Phase = sdk.WorkerPhaseUnknown
		}
	}
	// Determine the worker's start time based on pod start time
	if pod.Status.StartTime != nil {
		status.Started = &pod.Status.StartTime.Time
	}
	// Determine the worker's end time based on container[0] end time
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == pod.Spec.Containers[0].Name {
			if containerStatus.State.Terminated != nil {
				status.Ended =
					&pod.Status.ContainerStatuses[0].State.Terminated.FinishedAt.Time
			}
			break
		}
	}
	return status
}

// manageWorkerTimeout takes a pod and worker phase as input. If the phase is
// terminal and the timeout clock is already running for the pod, the clock is
// stopped. If the phase is NOT terminal and the timeout clock is NOT already
// running for the pod, the clock is started.
func (o *observer) manageWorkerTimeout(
	ctx context.Context,
	pod *corev1.Pod,
	phase sdk.WorkerPhase,
) {
	namespacedPodName := namespacedPodName(pod.Namespace, pod.Name)
	cancelFn, timed := o.timedPodsSet[namespacedPodName]
	if phase.IsTerminal() && timed {
		cancelFn() // Stop the clock
		return
	}
	if !phase.IsTerminal() && !timed {
		// Start the clock
		ctx, o.timedPodsSet[namespacedPodName] = context.WithCancel(ctx)
		go o.runWorkerTimerFn(ctx, pod)
	}
}

func (o *observer) runWorkerTimer(ctx context.Context, pod *corev1.Pod) {
	namespacedPodName := namespacedPodName(pod.Namespace, pod.Name)
	defer delete(o.timedPodsSet, namespacedPodName)
	timer := time.NewTimer(
		o.getPodTimeoutDuration(pod, o.config.maxWorkerLifetime),
	)
	defer timer.Stop()
	select {
	case <-timer.C:
		eventID := pod.Labels[myk8s.LabelEvent]
		// Create a new context for the timeout op. If we don't do this, the
		// possibility exists that the call to o.workersClient.Timeout() succeeds in
		// timing out the worker, but the worker is observed in its terminal,
		// timed-out state, resulting in cancelation of the current context before
		// the call to o.workersClient.Timeout() RETURNS, in which case we can end
		// up with an error telling us the context timed out during the
		// o.workersClient.Timeout(), when in fact, it has succeeded.
		timeoutCtx, cancel :=
			context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := o.workersClient.Timeout(timeoutCtx, eventID, nil); err != nil {
			o.errFn(
				errors.Wrapf(
					err,
					"error timing out worker for event %q",
					eventID,
				),
			)
		}
	case <-ctx.Done():
	}
}

func (o *observer) cleanupWorker(eventID string) {
	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	defer cancel()
	if err := o.workersClient.Cleanup(ctx, eventID, nil); err != nil {
		o.errFn(
			fmt.Sprintf(
				"error cleaning up after worker for event %q: %s",
				eventID,
				err,
			),
		)
	}
}
