package acid

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

func TestNewJobFromPod(t *testing.T) {
	now := time.Now()
	later := now.Add(time.Minute)
	podStartTime := metav1.NewTime(now)
	podEndTime := metav1.NewTime(later)
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testpod-abc123",
			Labels: map[string]string{
				"jobname": "testpod",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Image: "foo",
				},
			},
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
	expectedJob := &Job{
		ID:        "testpod-abc123",
		Name:      "testpod",
		Image:     "foo",
		StartTime: now,
		EndTime:   later,
		ExitCode:  0,
		Status:    JobSucceeded,
	}

	job := NewJobFromPod(pod)

	if !reflect.DeepEqual(job, expectedJob) {
		t.Errorf("job differs from expected job, got '%v', expected '%v'", job, expectedJob)
	}
}
