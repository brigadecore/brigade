package main

import (
	"sync"
	"testing"

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
