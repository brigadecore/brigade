package main

import (
	"sync"
	"testing"
	"time"

	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSyncDeletedPod(t *testing.T) {
	const testNamespace = "foo"
	const testPodName = "bar"
	observer := &observer{
		deletingPodsSet: map[string]struct{}{
			namespacedPodName(testNamespace, testPodName): {},
		},
		syncMu: &sync.Mutex{},
	}
	observer.syncDeletedPod(
		&corev1.Pod{
			ObjectMeta: v1.ObjectMeta{
				Namespace: testNamespace,
				Name:      testPodName,
			},
		},
	)
	require.Empty(t, observer.deletingPodsSet)
}

func TestGetPodTimeoutDuration(t *testing.T) {
	const maxTimeout = time.Duration(2)
	testCases := []struct {
		name     string
		pod      *corev1.Pod
		observer *observer
		expected time.Duration
	}{
		{
			name: "no duration annotation on pod",
			pod:  &corev1.Pod{},
			observer: &observer{
				errFn: func(i ...interface{}) {
					require.Fail(
						t,
						"errFn should not have been called, but was",
					)
				},
			},
			expected: maxTimeout,
		},
		{
			name: "duration annotation cannot be parsed",
			pod: &corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "1",
					},
				},
			},
			observer: &observer{
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					err, ok := i[0].(error)
					require.True(t, ok)
					require.Contains(t, err.Error(), "unable to parse timeout duration")
					require.Contains(t, err.Error(), "using configured maximum")
				},
			},
			expected: maxTimeout,
		},
		{
			name: "parsed duration exceeds the max",
			pod: &corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "3ns",
					},
				},
			},
			observer: &observer{
				errFn: func(i ...interface{}) {
					require.Len(t, i, 1)
					err, ok := i[0].(error)
					require.True(t, ok)
					require.Contains(t, err.Error(), "exceeds the configured maximum")
					require.Contains(t, err.Error(), "using configured maximum")
				},
			},
			expected: maxTimeout,
		},
		{
			name: "parsed duration does not exceed the max",
			pod: &corev1.Pod{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{
						myk8s.AnnotationTimeoutDuration: "1ns",
					},
				},
			},
			observer: &observer{
				errFn: func(i ...interface{}) {
					require.Fail(
						t,
						"errFn should not have been called, but was",
					)
				},
			},
			expected: time.Duration(1),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(
				t,
				testCase.expected,
				testCase.observer.getPodTimeoutDuration(testCase.pod, maxTimeout),
			)
		})
	}
}
