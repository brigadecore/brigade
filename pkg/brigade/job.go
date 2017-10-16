package brigade

import (
	"time"
)

// JobStatus is a label for the condition of a Job at the current time.
type JobStatus string

// These are the valid statuses of jobs.
const (
	// JobPending means the job has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	JobPending JobStatus = "Pending"
	// JobRunning means the job has been bound to a node and all of the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	JobRunning JobStatus = "Running"
	// JobSucceeded means that all containers in the job have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	JobSucceeded JobStatus = "Succeeded"
	// JobFailed means that all containers in the job have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	JobFailed JobStatus = "Failed"
	// JobUnknown means that for some reason the state of the job could not be obtained, typically due
	// to an error in communicating with the host of the job.
	JobUnknown JobStatus = "Unknown"
)

// Job is a single job that is executed when a build is triggered for an event.
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
