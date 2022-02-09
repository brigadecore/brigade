package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v3"
	coreTesting "github.com/brigadecore/brigade/sdk/v3/testing"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSyncJobPods(t *testing.T) {
	const testNamespace = "foo"
	const testPodName = "bar"

	var syncJobPodFnCallCount int
	mu := &sync.Mutex{}

	kubeClient := fake.NewSimpleClientset()

	observer := &observer{
		kubeClient: kubeClient,
		syncJobPodFn: func(_ interface{}) {
			mu.Lock()
			defer mu.Unlock()
			syncJobPodFnCallCount++
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	go observer.syncJobPods(ctx)

	// The informer needs a little time to get going. If we don't put a little
	// delay here, we'll be adding, updating, and deleting pods before the
	// informer gets cranking.
	<-time.After(time.Second)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: testPodName,
			Labels: map[string]string{
				myk8s.LabelComponent: myk8s.LabelKeyJob,
			},
		},
	}

	_, err := kubeClient.CoreV1().Pods(testNamespace).Create(
		ctx,
		pod,
		metav1.CreateOptions{},
	)
	require.NoError(t, err)

	_, err = kubeClient.CoreV1().Pods(testNamespace).Update(
		ctx,
		pod,
		metav1.UpdateOptions{},
	)
	require.NoError(t, err)

	err = kubeClient.CoreV1().Pods(testNamespace).Delete(
		ctx,
		testPodName,
		metav1.DeleteOptions{},
	)
	require.NoError(t, err)

	<-ctx.Done()

	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, 2, syncJobPodFnCallCount)
}

func TestSyncJobPod(t *testing.T) {
	now := time.Now().UTC()
	testCases := []struct {
		name     string
		pod      *corev1.Pod
		observer *observer
	}{
		{
			name: "pod is deleted",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: now,
					},
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{},
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.JobPhaseAborted, status.Phase)
						return nil
					},
				},
				cleanupJobFn: func(_, _ string) {},
			},
		},
		{
			name: "pod phase is pending",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{},
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.JobPhaseRunning, status.Phase)
						return nil
					},
				},
				cleanupJobFn: func(_, _ string) {
					require.Fail(
						t,
						"cleanupJobFn should not have been called, but was",
					)
				},
			},
		},
		{
			name: "pod phase is running and container[0] is not finished",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{},
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.JobPhaseRunning, status.Phase)
						require.Nil(t, status.Ended)
						return nil
					},
				},
				cleanupJobFn: func(_, _ string) {
					require.Fail(
						t,
						"cleanupJobFn should not have been called, but was",
					)
				},
			},
		},
		{
			name: "pod phase is running and container[0] succeeded",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "foo"}},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "foo",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 0,
									FinishedAt: metav1.Time{
										Time: now,
									},
								},
							},
						},
					},
				},
			},
			observer: &observer{
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.JobPhaseSucceeded, status.Phase)
						require.NotNil(t, status.Ended)
						require.Equal(t, now, *status.Ended)
						return nil
					},
				},
				cleanupJobFn: func(_, _ string) {},
			},
		},
		{
			name: "error updating job status",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{},
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						return errors.New("something went wrong")
					},
				},
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					str, ok := i[0].(string)
					require.True(t, ok)
					require.Contains(t, str, "something went wrong")
					require.Contains(t, str, "error updating status for event")
				},
			},
		},
		{
			name: "pod phase is running and container[0] failed",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "foo"}},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "foo",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
									ExitCode: 1,
									FinishedAt: metav1.Time{
										Time: now,
									},
								},
							},
						},
					},
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.JobPhaseFailed, status.Phase)
						require.NotNil(t, status.Ended)
						require.Equal(t, now, *status.Ended)
						return nil
					},
				},
				cleanupJobFn: func(_, _ string) {},
			},
		},
		{
			name: "pod phase is succeeded",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.JobPhaseSucceeded, status.Phase)
						return nil
					},
				},
				cleanupJobFn: func(_, _ string) {},
			},
		},
		{
			name: "pod phase is failed",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.JobPhaseFailed, status.Phase)
						return nil
					},
				},
				cleanupJobFn: func(_, _ string) {},
			},
		},
		{
			name: "pod phase is unknown",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodUnknown,
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				manageJobTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.JobPhase,
				) {
				},
				jobsClient: &coreTesting.MockJobsClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						jobName string,
						status sdk.JobStatus,
						_ *sdk.JobStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.JobPhaseUnknown, status.Phase)
						return nil
					},
				},
				cleanupJobFn: func(_, _ string) {
					require.Fail(
						t,
						"cleanupJobFn should not have been called, but was",
					)
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.observer.syncJobPod(testCase.pod)
		})
	}
}

func TestManageJobTimeout(t *testing.T) {
	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nombre",
			Namespace: "ns",
			Annotations: map[string]string{
				myk8s.AnnotationTimeoutDuration: "1m",
			},
		},
	}
	testCases := []struct {
		name       string
		phase      sdk.JobPhase
		observer   *observer
		assertions func(*observer)
	}{
		{
			name: "job in terminal phase and not already timed",
			// Nothing should happen
			phase: sdk.JobPhaseSucceeded,
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{},
			},
			assertions: func(o *observer) {
				require.Empty(t, o.timedPodsSet)
			},
		},
		{
			name: "job in terminal phase and already timed",
			// Should stop the clock
			phase: sdk.JobPhaseSucceeded,
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
			},
			assertions: func(o *observer) {
				require.Len(t, o.timedPodsSet, 1)
			},
		},
		{
			name: "job in non-terminal phase and not already timed",
			// Should start the clock
			phase: sdk.JobPhaseRunning,
			observer: &observer{
				timedPodsSet:  map[string]context.CancelFunc{},
				runJobTimerFn: func(context.Context, *corev1.Pod) {},
			},
			assertions: func(o *observer) {
				require.Contains(t, o.timedPodsSet, "ns:nombre")
			},
		},
		{
			name: "job in non-terminal phase and already timed",
			// Nothing should happen
			phase: sdk.JobPhaseRunning,
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
			},
			assertions: func(o *observer) {
				require.Contains(t, o.timedPodsSet, "ns:nombre")
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.observer.manageJobTimeout(
				context.Background(),
				testPod,
				testCase.phase,
			)
			testCase.assertions(testCase.observer)
		})
	}
}

func TestRunJobTimer(t *testing.T) {
	testCases := []struct {
		name       string
		pod        *corev1.Pod
		observer   *observer
		assertions func(*observer)
	}{
		{
			name: "canceled before timeout",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
					Labels: map[string]string{
						myk8s.LabelEvent: "tunguska",
						myk8s.LabelJob:   "italian",
					},
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "1m",
					},
				},
			},
			observer: &observer{
				config: observerConfig{
					maxJobLifetime: time.Minute,
				},
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				jobsClient: &coreTesting.MockJobsClient{
					TimeoutFn: func(
						context.Context,
						string,
						string,
						*sdk.JobTimeoutOptions,
					) error {
						require.Fail(
							t,
							"timout should not have been called on jobs client, but was",
						)
						return nil
					},
				},
				errFn: func(i ...interface{}) {
					require.Fail(t, "errFn should not have been called, but was")
				},
			},
			assertions: func(observer *observer) {
				require.Empty(t, observer.timedPodsSet)
			},
		},
		{
			name: "error calling timeout",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
					Labels: map[string]string{
						myk8s.LabelEvent: "tunguska",
						myk8s.LabelJob:   "italian",
					},
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "1s",
					},
				},
			},
			observer: &observer{
				config: observerConfig{
					maxJobLifetime: time.Minute,
				},
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				jobsClient: &coreTesting.MockJobsClient{
					TimeoutFn: func(
						context.Context,
						string,
						string,
						*sdk.JobTimeoutOptions,
					) error {
						return errors.New("something went wrong")
					},
				},
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					err, ok := i[0].(error)
					require.True(t, ok)
					require.Contains(t, err.Error(), "something went wrong")
				},
			},
			assertions: func(observer *observer) {
				require.Empty(t, observer.timedPodsSet)
			},
		},
		{
			name: "success",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
					Labels: map[string]string{
						myk8s.LabelEvent: "tunguska",
						myk8s.LabelJob:   "italian",
					},
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "1s",
					},
				},
			},
			observer: &observer{
				config: observerConfig{
					maxJobLifetime: time.Minute,
				},
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				jobsClient: &coreTesting.MockJobsClient{
					TimeoutFn: func(
						context.Context,
						string,
						string,
						*sdk.JobTimeoutOptions,
					) error {
						return nil
					},
				},
				errFn: func(i ...interface{}) {
					require.Fail(t, "errFn should not have been called, but was")
				},
			},
			assertions: func(observer *observer) {
				require.Empty(t, observer.timedPodsSet)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			func() {
				// A context that's longer than the timeout of 1s
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
				testCase.observer.runJobTimer(ctx, testCase.pod)
			}()
		})
	}
}

func TestCleanupJob(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	testCases := []struct {
		name     string
		observer *observer
	}{
		{
			name: "error calling cleanup",
			observer: &observer{
				config: observerConfig{
					delayBeforeCleanup: time.Second,
				},
				jobsClient: &coreTesting.MockJobsClient{
					CleanupFn: func(
						context.Context,
						string,
						string,
						*sdk.JobCleanupOptions,
					) error {
						return errors.New("something went wrong")
					},
				},
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					msg, ok := i[0].(string)
					require.True(t, ok)
					require.Contains(t, msg, "something went wrong")
					require.Contains(t, msg, "error cleaning up after event")
				},
			},
		},
		{
			name: "success",
			observer: &observer{
				config: observerConfig{
					delayBeforeCleanup: time.Second,
				},
				jobsClient: &coreTesting.MockJobsClient{
					CleanupFn: func(
						context.Context,
						string,
						string,
						*sdk.JobCleanupOptions,
					) error {
						return nil
					},
				},
				errFn: func(i ...interface{}) {
					require.Fail(
						t,
						"error logging function should not have been called",
					)
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.observer.cleanupJob(testEventID, testJobName)
		})
	}
}
