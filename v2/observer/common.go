package main

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

const apiRequestTimeout = 30 * time.Second

// syncDeletedPod only fires when a pod deletion is COMPLETE. i.e. The pod is
// completely gone.
func (o *observer) syncDeletedPod(obj interface{}) {
	o.syncMu.Lock()
	defer o.syncMu.Unlock()
	pod := obj.(*corev1.Pod)
	// Remove this pod from the set of pods we were tracking for deletion.
	// Managing this set is essential to not leaking memory.
	delete(o.deletingPodsSet, namespacedPodName(pod.Namespace, pod.Name))
}
