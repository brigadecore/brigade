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

func TestSyncWorkerPods(t *testing.T) {
	const testNamespace = "foo"
	const testPodName = "bar"

	var syncWorkerPodFnCallCount int
	mu := &sync.Mutex{}

	kubeClient := fake.NewSimpleClientset()

	observer := &observer{
		kubeClient: kubeClient,
		syncWorkerPodFn: func(_ interface{}) {
			mu.Lock()
			defer mu.Unlock()
			syncWorkerPodFnCallCount++
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	go observer.syncWorkerPods(ctx)

	// The informer needs a little time to get going. If we don't put a little
	// delay here, we'll be adding, updating, and deleting pods before the
	// informer gets cranking.
	<-time.After(time.Second)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: testPodName,
			Labels: map[string]string{
				myk8s.LabelComponent: myk8s.LabelKeyWorker,
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
	require.Equal(t, 2, syncWorkerPodFnCallCount)
}

func TestSyncWorkerPod(t *testing.T) {
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
				manageWorkerTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.WorkerPhase,
				) {
				},
				workersClient: &coreTesting.MockWorkersClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						status sdk.WorkerStatus,
						_ *sdk.WorkerStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.WorkerPhaseAborted, status.Phase)
						return nil
					},
				},
				cleanupWorkerFn: func(string) {},
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
				manageWorkerTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.WorkerPhase,
				) {
				},
				workersClient: &coreTesting.MockWorkersClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						status sdk.WorkerStatus,
						_ *sdk.WorkerStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.WorkerPhaseRunning, status.Phase)
						return nil
					},
				},
				cleanupWorkerFn: func(string) {
					require.Fail(
						t,
						"cleanupWorkerFn should not have been called, but was",
					)
				},
			},
		},
		{
			name: "pod phase is running",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					StartTime: &metav1.Time{
						Time: now,
					},
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{},
				manageWorkerTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.WorkerPhase,
				) {
				},
				workersClient: &coreTesting.MockWorkersClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						status sdk.WorkerStatus,
						_ *sdk.WorkerStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.WorkerPhaseRunning, status.Phase)
						require.NotNil(t, now, status.Started)
						require.Equal(t, now, *status.Started)
						return nil
					},
				},
				cleanupWorkerFn: func(string) {
					require.Fail(
						t,
						"cleanupWorkerFn should not have been called, but was",
					)
				},
			},
		},
		{
			name: "pod phase is succeeded",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "foo",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
					StartTime: &metav1.Time{
						Time: now,
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "foo",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
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
					"ns/nombre": func() {},
				},
				manageWorkerTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.WorkerPhase,
				) {
				},
				workersClient: &coreTesting.MockWorkersClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						status sdk.WorkerStatus,
						_ *sdk.WorkerStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.WorkerPhaseSucceeded, status.Phase)
						require.NotNil(t, now, status.Started)
						require.Equal(t, now, *status.Started)
						require.NotNil(t, now, status.Ended)
						require.Equal(t, now, *status.Ended)
						return nil
					},
				},
				cleanupWorkerFn: func(string) {},
			},
		},
		{
			name: "pod phase is failed",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nombre",
					Namespace: "ns",
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "foo",
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
					StartTime: &metav1.Time{
						Time: now,
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{
							Name: "foo",
							State: corev1.ContainerState{
								Terminated: &corev1.ContainerStateTerminated{
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
					"ns/nombre": func() {},
				},
				manageWorkerTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.WorkerPhase,
				) {
				},
				workersClient: &coreTesting.MockWorkersClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						status sdk.WorkerStatus,
						_ *sdk.WorkerStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.WorkerPhaseFailed, status.Phase)
						require.NotNil(t, now, status.Started)
						require.Equal(t, now, *status.Started)
						require.NotNil(t, now, status.Ended)
						require.Equal(t, now, *status.Ended)
						return nil
					},
				},
				cleanupWorkerFn: func(string) {},
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
					"ns/nombre": func() {},
				},
				manageWorkerTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.WorkerPhase,
				) {
				},
				workersClient: &coreTesting.MockWorkersClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						status sdk.WorkerStatus,
						_ *sdk.WorkerStatusUpdateOptions,
					) error {
						require.Equal(t, sdk.WorkerPhaseUnknown, status.Phase)
						return nil
					},
				},
				cleanupWorkerFn: func(string) {
					require.Fail(
						t,
						"cleanupWorkerFn should not have been called, but was",
					)
				},
			},
		},
		{
			name: "error updating worker status",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{},
				manageWorkerTimeoutFn: func(
					context.Context,
					*corev1.Pod,
					sdk.WorkerPhase,
				) {
				},
				workersClient: &coreTesting.MockWorkersClient{
					UpdateStatusFn: func(
						ctx context.Context,
						eventID string,
						status sdk.WorkerStatus,
						_ *sdk.WorkerStatusUpdateOptions,
					) error {
						return errors.New("something went wrong")
					},
				},
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					require.Contains(t, i[0].(string), "something went wrong")
					require.Contains(t, i[0].(string), "error updating status for event")
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.observer.syncWorkerPod(testCase.pod)
		})
	}
}

func TestManageWorkerTimeout(t *testing.T) {
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
		phase      sdk.WorkerPhase
		observer   *observer
		assertions func(*observer)
	}{
		{
			name: "worker in terminal phase and not already timed",
			// Nothing should happen
			phase: sdk.WorkerPhaseSucceeded,
			observer: &observer{
				timedPodsSet: map[string]context.CancelFunc{},
			},
			assertions: func(o *observer) {
				require.Empty(t, o.timedPodsSet)
			},
		},
		{
			name: "worker in terminal phase and already timed",
			// Should stop the clock
			phase: sdk.WorkerPhaseSucceeded,
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
			name: "worker in non-terminal phase and not already timed",
			// Should start the clock
			phase: sdk.WorkerPhaseRunning,
			observer: &observer{
				timedPodsSet:     map[string]context.CancelFunc{},
				runWorkerTimerFn: func(context.Context, *corev1.Pod) {},
			},
			assertions: func(o *observer) {
				require.Contains(t, o.timedPodsSet, "ns:nombre")
			},
		},
		{
			name: "worker in non-terminal phase and already timed",
			// Nothing should happen
			phase: sdk.WorkerPhaseRunning,
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
			testCase.observer.manageWorkerTimeout(
				context.Background(),
				testPod,
				testCase.phase,
			)
			testCase.assertions(testCase.observer)
		})
	}
}

func TestRunWorkerTimer(t *testing.T) {
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
					},
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "1m",
					},
				},
			},
			observer: &observer{
				config: observerConfig{
					maxWorkerLifetime: time.Minute,
				},
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				workersClient: &coreTesting.MockWorkersClient{
					TimeoutFn: func(
						context.Context,
						string,
						*sdk.WorkerTimeoutOptions,
					) error {
						require.Fail(
							t,
							"timout should not have been called on workers client, but was",
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
					},
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "1s",
					},
				},
			},
			observer: &observer{
				config: observerConfig{
					maxWorkerLifetime: time.Minute,
				},
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				workersClient: &coreTesting.MockWorkersClient{
					TimeoutFn: func(
						context.Context,
						string,
						*sdk.WorkerTimeoutOptions,
					) error {
						return errors.New("something went wrong")
					},
				},
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					require.Contains(t, i[0].(error).Error(), "something went wrong")
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
					},
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "1s",
					},
				},
			},
			observer: &observer{
				config: observerConfig{
					maxWorkerLifetime: time.Minute,
				},
				timedPodsSet: map[string]context.CancelFunc{
					"ns:nombre": func() {},
				},
				workersClient: &coreTesting.MockWorkersClient{
					TimeoutFn: func(
						context.Context,
						string,
						*sdk.WorkerTimeoutOptions,
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
				testCase.observer.runWorkerTimer(ctx, testCase.pod)
				testCase.assertions(testCase.observer)
			}()
		})
	}
}

func TestCleanupWorker(t *testing.T) {
	const testEventID = "123456789"
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
				workersClient: &coreTesting.MockWorkersClient{
					CleanupFn: func(
						context.Context,
						string,
						*sdk.WorkerCleanupOptions,
					) error {
						return errors.New("something went wrong")
					},
				},
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					msg, ok := i[0].(string)
					require.True(t, ok)
					require.Contains(t, msg, "something went wrong")
					require.Contains(t, msg, "error cleaning up after worker for event")
				},
			},
		},
		{
			name: "success",
			observer: &observer{
				config: observerConfig{
					delayBeforeCleanup: time.Second,
				},
				workersClient: &coreTesting.MockWorkersClient{
					CleanupFn: func(
						context.Context,
						string,
						*sdk.WorkerCleanupOptions,
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
			testCase.observer.cleanupWorker(testEventID)
		})
	}
}
