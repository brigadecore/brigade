package worker

import (
	"github.com/deis/acid/pkg/k8s"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
)

// k8sExecutor is a Kubernetes executor.
type k8sExecutor struct {
	client kubernetes.Interface
}

func (s *k8sExecutor) Create(namespace string, pod *v1.Pod) (*v1.Pod, error) {
	// We do a lazy connection so that the constructor does not return an
	// error.
	if s.client == nil {
		// creates the in-cluster config
		// creates the clientset
		clientset, err := k8s.Client()
		if err != nil {
			return pod, err
		}
		s.client = clientset
	}
	return s.client.CoreV1().Pods(namespace).Create(pod)
}
