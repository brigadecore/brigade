package main

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (o *observer) syncJobPods(ctx context.Context) {
	jobsSelector := myk8s.JobPodsSelector(o.config.brigadeID)
	jobPodsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = jobsSelector
				return o.kubeClient.CoreV1().Pods("").List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = jobsSelector
				return o.kubeClient.CoreV1().Pods("").Watch(ctx, options)
			},
		},
		&corev1.Pod{},
		0,
		cache.Indexers{},
	)
	jobPodsInformer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: o.syncJobPodFn,
			UpdateFunc: func(_, newObj interface{}) {
				o.syncJobPodFn(newObj)
			},
		},
	)
	jobPodsInformer.Run(ctx.Done())
}

func (o *observer) syncJobPod(obj interface{}) {
	ctx := context.Background()
	pod := obj.(*corev1.Pod) // nolint: forcetypeassert
	// Map pod status to job status
	status := o.getJobStatusFromPod(pod)
	// Manage the timeout clock
	o.manageJobTimeoutFn(ctx, pod, status.Phase)
	// Use the API to update Job status
	eventID := pod.Labels[myk8s.LabelEvent]
	jobName := pod.Labels[myk8s.LabelJob]
	ctx, cancel := context.WithTimeout(ctx, apiRequestTimeout)
	defer cancel()
	if err := o.jobsClient.UpdateStatus(
		ctx,
		eventID,
		jobName,
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
					"error updating status for event %q worker job %q: %s",
					eventID,
					jobName,
					err,
				),
			)
		}
	}
	if pod.DeletionTimestamp != nil {
		// If the pod was deleted, immediately complete cleanup of other resources
		// associated with the job
		o.cleanupJobFn(eventID, jobName)
	} else if status.Phase.IsTerminal() {
		// Otherwise, if the job is in a terminal phase, defer cleanup so the log
		// agent has a chance to catch up if necessary.
		go func() {
			<-time.After(o.config.delayBeforeCleanup)
			o.cleanupJobFn(eventID, jobName)
		}()
	}
}

func (o *observer) getJobStatusFromPod(pod *corev1.Pod) core.JobStatus {
	// Determine the jobs's phase based on pod phase
	status := core.JobStatus{
		Phase: core.JobPhaseRunning,
	}
	if pod.DeletionTimestamp != nil {
		// This pod has been deleted. Pods usually hang around for a while after
		// they complete before they are deleted (to ensure the log agent has enough
		// time to capture all logs), so the job's final transition to a terminal
		// phase has PROBABLY been recorded already, BUT given the possibility that
		// the job's pod was manually deleted by an operator, we'll attempt a final
		// phase change to ABORTED here, knowing that it will simply fail if the job
		// has already reached a terminal state.
		status.Phase = core.JobPhaseAborted
	} else {
		// Otherwise, the job's phase depends on the pod's phase
		switch pod.Status.Phase {
		case corev1.PodPending:
			// For Brigade's purposes, this counts as running
			status.Phase = core.JobPhaseRunning
			// Unless... when an image pull backoff occurs, the pod still shows as
			// pending. We account for that here and treat it as a failure.
			//
			// TODO: Are there other conditions we need to watch out for?
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil &&
					(containerStatus.State.Waiting.Reason == "ImagePullBackOff" ||
						containerStatus.State.Waiting.Reason == "ErrImagePull") {
					status.Phase = core.JobPhaseFailed
					break
				}
			}
		case corev1.PodRunning:
			status.Phase = core.JobPhaseRunning
		case corev1.PodSucceeded:
			status.Phase = core.JobPhaseSucceeded
		case corev1.PodFailed:
			status.Phase = core.JobPhaseFailed
		case corev1.PodUnknown:
			status.Phase = core.JobPhaseUnknown
		}
	}
	// Due to our sidecar support, even if the phase is running, we're not REALLY
	// running if container[0] has exited. Adjust accordingly.
	if status.Phase == core.JobPhaseRunning {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name == pod.Spec.Containers[0].Name {
				if containerStatus.State.Terminated != nil {
					if containerStatus.State.Terminated.ExitCode == 0 {
						status.Phase = core.JobPhaseSucceeded
					} else {
						status.Phase = core.JobPhaseFailed
					}
				}
				break
			}
		}
	}
	// Determine the job's start time based on pod start time
	if pod.Status.StartTime != nil {
		status.Started = &pod.Status.StartTime.Time
	}
	// Determine the job's end time based on container[0] end time
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == pod.Spec.Containers[0].Name {
			if containerStatus.State.Terminated != nil {
				status.Ended = &containerStatus.State.Terminated.FinishedAt.Time
			}
			break
		}
	}
	return status
}

// manageJobTimeout takes a pod and job phase as input. If the phase is
// terminal and the timeout clock is already running for the pod, the clock is
// stopped. If the phase is NOT terminal and the timeout clock is NOT already
// running for the pod, the clock is started.
func (o *observer) manageJobTimeout(
	ctx context.Context,
	pod *corev1.Pod,
	phase core.JobPhase,
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
		go o.runJobTimerFn(ctx, pod)
	}
}

func (o *observer) runJobTimer(ctx context.Context, pod *corev1.Pod) {
	namespacedPodName := namespacedPodName(pod.Namespace, pod.Name)
	defer delete(o.timedPodsSet, namespacedPodName)
	timer := time.NewTimer(
		o.getPodTimeoutDuration(pod, o.config.maxJobLifetime),
	)
	defer timer.Stop()
	select {
	case <-timer.C:
		eventID := pod.Labels[myk8s.LabelEvent]
		jobName := pod.Labels[myk8s.LabelJob]
		// Create a new context for the timeout op. If we don't do this, the
		// possibility exists that the call to o.jobsClient.Timeout() succeeds in
		// timing out the job, but the job is observed in its terminal, timed-out
		// state, resulting in cancelation of the current context before the call to
		// o.jobsClient.Timeout() RETURNS, in which case we can end up with an error
		// telling us the context timed out during the o.jobsClient.Timeout(), when
		// in fact, it has succeeded.
		timeoutCtx, cancel :=
			context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err :=
			o.jobsClient.Timeout(timeoutCtx, eventID, jobName, nil); err != nil {
			o.errFn(
				errors.Wrapf(
					err,
					"error timing out job %q for event %q",
					jobName,
					eventID,
				),
			)
		}
	case <-ctx.Done():
	}
}

func (o *observer) cleanupJob(eventID string, jobName string) {
	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	defer cancel()
	if err := o.jobsClient.Cleanup(ctx, eventID, jobName, nil); err != nil {
		o.errFn(
			fmt.Sprintf(
				"error cleaning up after event %q job %q: %s",
				eventID,
				jobName,
				err,
			),
		)
	}
}
