package kube

import (
	"fmt"

	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/deis/acid/pkg/acid"
)

// GetWorker returns the worker description.
//
// This will return an error if no worker is found for the build, which can
// happen when a build is scheduled, but not yet started.
func (s *store) GetWorker(buildID string) (*acid.Worker, error) {
	labels := labels.Set{"heritage": "acid", "build": buildID}
	listOption := meta.ListOptions{LabelSelector: labels.AsSelector().String()}
	pods, err := s.client.CoreV1().Pods(s.namespace).List(listOption)
	if err != nil {
		return nil, err
	}
	if len(pods.Items) < 1 {
		return nil, fmt.Errorf("could not find worker for build %s: no pod exists with label %s", buildID, labels.AsSelector().String())
	}
	return NewWorkerFromPod(pods.Items[0]), nil
}

// NewWorkerFromPod creates a new *Worker from a pod definition.
func NewWorkerFromPod(pod v1.Pod) *acid.Worker {
	l := pod.Labels
	worker := &acid.Worker{
		ID:        pod.Name,
		BuildID:   l["build"],
		ProjectID: l["project"],
		Commit:    l["commit"],
		Status:    acid.JobStatus(pod.Status.Phase),
	}

	if (worker.Status != acid.JobPending) && (worker.Status != acid.JobUnknown) {
		worker.StartTime = pod.Status.StartTime.Time
	}

	if len(pod.Status.ContainerStatuses) > 0 {
		cs := pod.Status.ContainerStatuses[0]
		if cs.State.Terminated != nil {
			worker.EndTime = cs.State.Terminated.FinishedAt.Time
			worker.ExitCode = cs.State.Terminated.ExitCode
		}
	}

	return worker
}
