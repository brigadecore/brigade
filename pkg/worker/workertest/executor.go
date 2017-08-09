package workertest

import (
	"k8s.io/client-go/pkg/api/v1"
)

type MockExecutor struct {
	LastPod *v1.Pod
}

func (s *MockExecutor) Create(ns string, pod *v1.Pod) (*v1.Pod, error) {
	s.LastPod = pod
	return pod, nil
}
