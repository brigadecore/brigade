package kube

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/Azure/brigade/pkg/brigade"
)

func TestNewWorkerFromPod(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Minute)
	expect := &brigade.Worker{
		ID:        "pod-name",
		BuildID:   "build-id",
		ProjectID: "project-id",
		Commit:    "commit-sha",
		StartTime: now,
		EndTime:   later,
		ExitCode:  0,
		Status:    brigade.JobSucceeded,
	}

	podStartTime := metav1.NewTime(now)
	podEndTime := metav1.NewTime(later)
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: expect.ID,
			Labels: map[string]string{
				"build":   expect.BuildID,
				"project": expect.ProjectID,
				"commit":  expect.Commit,
			},
			CreationTimestamp: podStartTime,
		},
		Status: v1.PodStatus{
			Phase:     v1.PodSucceeded,
			StartTime: &podStartTime,
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode:   0,
							FinishedAt: podEndTime,
						},
					},
				},
			},
		},
	}

	worker := NewWorkerFromPod(pod)

	if !reflect.DeepEqual(worker, expect) {
		t.Errorf("worker differs from expected, got '%v', expected '%v'", worker, expect)
	}
}
