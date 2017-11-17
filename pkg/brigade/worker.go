package brigade

import "time"

// Worker represents the worker that runs a build.
// A worker executes (and wraps) the jobs in a build.
type Worker struct {
	// ID is the name for the pod running this job
	ID string `json:"id"`
	// BuildID is the build ID (ULID).
	BuildID string `json:"build_id"`
	// ProjectID is the computed name of the project (brigade-aeff2343a3234ff)
	ProjectID string `json:"project_id"`
	// StartTime is the time the worker started.
	StartTime time.Time `json:"start_time"`
	// EndTime is the time the worker completed. This may not be present
	// if the job has not completed.
	EndTime time.Time `json:"end_time"`
	// ExitCode is the exit code of the job. This may not be present
	// if the job has not completed.
	ExitCode int32 `json:"exit_code"`
	// Status is a textual representation of the job's running status
	Status JobStatus `json:"status"`
}
