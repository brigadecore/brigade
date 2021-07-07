package main

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func (o *observer) syncJobPods(ctx context.Context) {
	jobPodsInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = myk8s.JobPodsSelector()
				return o.kubeClient.CoreV1().Pods("").List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = myk8s.JobPodsSelector()
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
			DeleteFunc: o.syncDeletedPodFn,
		},
	)
	jobPodsInformer.Run(ctx.Done())
}

// nolint: gocyclo
func (o *observer) syncJobPod(obj interface{}) {
	pod := obj.(*corev1.Pod)

	// Fetch the cancel func associated with a timed job pod
	lookup := namespacedPodName(pod.Namespace, pod.Name)
	cancelTimer, exists := o.timedPodsSet[lookup]
	if !exists {
		var timerCtx context.Context
		timerCtx, cancelTimer = context.WithCancel(context.Background())
		o.timedPodsSet[lookup] = cancelTimer
		o.startJobPodTimerFn(timerCtx, pod)
	}

	// Job pods are only deleted after we're FULLY done with them. So if the
	// DeletionTimestamp is set, there's nothing for us to do because the Job must
	// already be in a terminal state.
	if pod.DeletionTimestamp != nil {
		return
	}

	status := core.JobStatus{
		Phase: core.JobPhaseRunning,
	}
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
		cancelTimer()
	case corev1.PodFailed:
		status.Phase = core.JobPhaseFailed
		cancelTimer()
	case corev1.PodUnknown:
		status.Phase = core.JobPhaseUnknown
		cancelTimer()
	}

	if pod.Status.StartTime != nil {
		status.Started = &pod.Status.StartTime.Time
	}
	// Pods don't really have an end time. We grab the end time of container[0]
	// because that's what we really care about.
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Name == pod.Spec.Containers[0].Name {
			if containerStatus.State.Terminated != nil {
				status.Ended = &containerStatus.State.Terminated.FinishedAt.Time
			}
			break
		}
	}

	// If the phase is running, we're not REALLY running if container[0] has
	// exited. Adjust accordingly.
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

	// Use the API to update Job status
	eventID := pod.Labels[myk8s.LabelEvent]
	jobName := pod.Labels[myk8s.LabelJob]
	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	defer cancel()
	if err := o.jobsClient.UpdateStatus(
		ctx,
		eventID,
		jobName,
		status,
	); err != nil {
		o.errFn(
			fmt.Sprintf(
				"error updating status for event %q worker job %q: %s",
				eventID,
				jobName,
				err,
			),
		)
	}

	if status.Phase == core.JobPhaseSucceeded ||
		status.Phase == core.JobPhaseFailed {
		go o.deleteJobResourcesFn(pod.Namespace, pod.Name, eventID, jobName)
	}
}

// deleteJobResources deletes a Job pod after a 60 second delay. The delay is to
// ensure any log aggregators have a chance to get all logs from the completed
// pod before it is torpedoed.
func (o *observer) deleteJobResources(
	namespace string,
	podName string,
	eventID string,
	jobName string,
) {
	namespacedJobPodName := namespacedPodName(namespace, podName)

	o.syncMu.Lock()
	if _, alreadyDeleting :=
		o.deletingPodsSet[namespacedJobPodName]; alreadyDeleting {
		o.syncMu.Unlock()
		return
	}
	o.deletingPodsSet[namespacedJobPodName] = struct{}{}
	o.syncMu.Unlock()

	<-time.After(o.config.delayBeforeCleanup)

	ctx, cancel := context.WithTimeout(context.Background(), apiRequestTimeout)
	defer cancel()
	if err := o.jobsClient.Cleanup(ctx, eventID, jobName); err != nil {
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

// startJobPodTimer inspects the provided Job pod for a timeoutDuration
// annotation value and if non-empty, starts a timer using the parsed value.
// If the timeout is reached, we make an API call to execute the appropriate
// logic. Alternatively, the context may be canceled in the meantime, which
// will stop the timer.
func (o *observer) startJobPodTimer(ctx context.Context, pod *corev1.Pod) {
	defer delete(o.timedPodsSet, namespacedPodName(pod.Namespace, pod.Name))

	if pod.Status.Phase != corev1.PodPending &&
		pod.Status.Phase != corev1.PodRunning {
		return
	}

	go func() {
		timer := time.NewTimer(
			o.getPodTimeoutDuration(pod, o.config.maxJobLifetime),
		)
		defer timer.Stop()

		select {
		case <-timer.C:
			eventID := pod.Labels[myk8s.LabelEvent]
			jobName := pod.Labels[myk8s.LabelJob]
			if err := o.jobsClient.Timeout(ctx, eventID, jobName); err != nil {
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
	}()
}
