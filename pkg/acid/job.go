package acid

import (
	"time"

	"k8s.io/client-go/pkg/api/v1"
)

type JobStatus v1.PodPhase

// These are the valid statuses of jobs.
const (
	// JobPending means the job has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	JobPending JobStatus = JobStatus(v1.PodPending)
	// JobRunning means the job has been bound to a node and all of the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	JobRunning JobStatus = JobStatus(v1.PodRunning)
	// JobSucceeded means that all containers in the job have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	JobSucceeded JobStatus = JobStatus(v1.PodSucceeded)
	// JobFailed means that all containers in the job have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	JobFailed JobStatus = JobStatus(v1.PodFailed)
	// JobUnknown means that for some reason the state of the job could not be obtained, typically due
	// to an error in communicating with the host of the job.
	JobUnknown JobStatus = JobStatus(v1.PodUnknown)
)

type Job struct {
	// ID is the name for the pod running this job
	ID string `json:"id"`
	// Name is the name for the job
	Name string `json:"name"`
	// Image is the execution environment running the job
	Image string `json:"image"`
	// CreationTime is a timestamp representing the server time when this object was
	// created. It is not guaranteed to be set in happens-before order across separate operations.
	CreationTime time.Time `json:"creation_time"`
	// StartTime is the time the job started.
	StartTime time.Time `json:"start_time"`
	// EndTime is the time the job completed. This may not be present
	// if the job has not completed.
	EndTime time.Time `json:"end_time"`
	// ExitCode is the exit code of the job. This may not be present
	// if the job has not completed.
	ExitCode int32 `json:"exit_code"`
	// Status is a textual representation of the job's running status
	Status JobStatus `json:"status"`
}

// NewJobFromPod parses the pod object metadata and deserializes it into a Job.
func NewJobFromPod(pod v1.Pod) *Job {
	job := &Job{
		ID:           pod.ObjectMeta.Name,
		Name:         pod.ObjectMeta.Labels["jobname"],
		CreationTime: pod.ObjectMeta.CreationTimestamp.Time,
		Image:        pod.Spec.Containers[0].Image,
		Status:       JobStatus(pod.Status.Phase),
	}

	if (job.Status != JobPending) && (job.Status != JobUnknown) {
		job.StartTime = pod.Status.StartTime.Time
	}

	if len(pod.Status.ContainerStatuses) > 0 {
		if pod.Status.ContainerStatuses[0].State.Terminated != nil {
			job.EndTime = pod.Status.ContainerStatuses[0].State.Terminated.FinishedAt.Time
			job.ExitCode = pod.Status.ContainerStatuses[0].State.Terminated.ExitCode
		}
	}

	return job
}
