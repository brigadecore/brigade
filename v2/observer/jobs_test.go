package main

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2/core"
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
	var syncDeletedPodFnCalled bool
	mu := &sync.Mutex{}

	kubeClient := fake.NewSimpleClientset()

	observer := &observer{
		kubeClient: kubeClient,
		syncJobPodFn: func(_ interface{}) {
			mu.Lock()
			defer mu.Unlock()
			syncJobPodFnCallCount++
		},
		syncDeletedPodFn: func(_ interface{}) {
			mu.Lock()
			defer mu.Unlock()
			syncDeletedPodFnCalled = true
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
	require.True(t, syncDeletedPodFnCalled)
}

func TestSyncJobPod(t *testing.T) {
	now := time.Now().UTC()
	testCases := []struct {
		name     string
		pod      *corev1.Pod
		observer *observer
	}{
		{
			name: "deletionTimestamp is not nil",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: now,
					},
				},
			},
			observer: &observer{
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					require.Fail(
						t,
						"updateJobStatusFn should not have been called, but was",
					)
					return nil
				},
				deleteJobResourcesFn: func(_, _, _, _ string) {
					require.Fail(
						t,
						"deleteJobResourcesFn should not have been called, but was",
					)
				},
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
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					require.Equal(t, core.JobPhaseRunning, status.Phase)
					return nil
				},
				deleteJobResourcesFn: func(_, _, _, _ string) {
					require.Fail(
						t,
						"deleteJobResourcesFn should not have been called, but was",
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
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					require.Equal(t, core.JobPhaseRunning, status.Phase)
					require.Nil(t, status.Ended)
					return nil
				},
				deleteJobResourcesFn: func(_, _, _, _ string) {
					require.Fail(
						t,
						"deleteJobResourcesFn should not have been called, but was",
					)
				},
			},
		},
		{
			name: "pod phase is running and container[0] succeeded",
			pod: &corev1.Pod{
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
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					require.Equal(t, core.JobPhaseSucceeded, status.Phase)
					require.NotNil(t, status.Ended)
					require.Equal(t, now, *status.Ended)
					return nil
				},
				deleteJobResourcesFn: func(_, _, _, _ string) {},
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
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					return errors.New("something went wrong")
				},
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					require.Contains(t, i[0].(string), "something went wrong")
					require.Contains(t, i[0].(string), "error updating status for event")
				},
			},
		},
		{
			name: "pod phase is running and container[0] failed",
			pod: &corev1.Pod{
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
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					require.Equal(t, core.JobPhaseFailed, status.Phase)
					require.NotNil(t, status.Ended)
					require.Equal(t, now, *status.Ended)
					return nil
				},
				deleteJobResourcesFn: func(_, _, _, _ string) {},
			},
		},
		{
			name: "pod phase is succeeded",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodSucceeded,
				},
			},
			observer: &observer{
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					require.Equal(t, core.JobPhaseSucceeded, status.Phase)
					return nil
				},
				deleteJobResourcesFn: func(_, _, _, _ string) {},
			},
		},
		{
			name: "pod phase is failed",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodFailed,
				},
			},
			observer: &observer{
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					require.Equal(t, core.JobPhaseFailed, status.Phase)
					return nil
				},
				deleteJobResourcesFn: func(_, _, _, _ string) {},
			},
		},
		{
			name: "pod phase is unknown",
			pod: &corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodUnknown,
				},
			},
			observer: &observer{
				updateJobStatusFn: func(
					ctx context.Context,
					eventID string,
					jobName string,
					status core.JobStatus,
				) error {
					require.Equal(t, core.JobPhaseUnknown, status.Phase)
					return nil
				},
				deleteJobResourcesFn: func(_, _, _, _ string) {
					require.Fail(
						t,
						"deleteJobResourcesFn should not have been called, but was",
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

func TestDeleteJobResources(t *testing.T) {
	const testNamespace = "foo"
	const testPodName = "bar"
	const testEventID = "123456789"
	const testJobName = "italian"
	testCases := []struct {
		name     string
		observer *observer
	}{
		{
			name: "already tracking delete",
			observer: &observer{
				deletingPodsSet: map[string]struct{}{
					namespacedPodName(testNamespace, testPodName): {},
				},
				syncMu: &sync.Mutex{},
				errFn: func(i ...interface{}) {
					require.Fail(
						t,
						"error logging function should not have been called",
					)
				},
			},
		},
		{
			name: "error calling cleanup",
			observer: &observer{
				config: observerConfig{
					delayBeforeCleanup: time.Second,
				},
				deletingPodsSet: map[string]struct{}{},
				syncMu:          &sync.Mutex{},
				cleanupJobFn: func(context.Context, string, string) error {
					return errors.New("something went wrong")
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
				deletingPodsSet: map[string]struct{}{},
				syncMu:          &sync.Mutex{},
				cleanupJobFn: func(context.Context, string, string) error {
					return nil
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
			testCase.observer.deleteJobResources(
				testNamespace,
				testPodName,
				testEventID,
				testJobName,
			)
		})
	}
}
