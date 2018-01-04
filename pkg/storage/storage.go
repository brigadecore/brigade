package storage

import (
	"io"

	"github.com/Azure/brigade/pkg/brigade"
)

// Store represents a storage engine for a Project.
type Store interface {
	// GetProjects retrieves all projects from storage.
	GetProjects() ([]*brigade.Project, error)
	// GetProject retrieves the project from storage.
	GetProject(id string) (*brigade.Project, error)
	// GetProjectBuilds retrieves the project's builds from storage.
	GetProjectBuilds(proj *brigade.Project) ([]*brigade.Build, error)
	// GetBuilds retrieves all active builds from storage.
	GetBuilds() ([]*brigade.Build, error)
	// GetBuild retrieves the build from storage.
	GetBuild(id string) (*brigade.Build, error)
	// CreateBuild creates a new job for the work queue.
	CreateBuild(build *brigade.Build) error
	// GetBuildJobs retrieves all build jobs (pods) from storage.
	GetBuildJobs(build *brigade.Build) ([]*brigade.Job, error)
	// GetWorker returns the worker for a given build.
	GetWorker(buildID string) (*brigade.Worker, error)
	// GetJob retrieves the job from storage.
	GetJob(id string) (*brigade.Job, error)
	// GetJobLog retrieves all logs for a job from storage.
	GetJobLog(job *brigade.Job) (string, error)
	// GetJobLogStream retrieve a stream of all logs for a job from storage.
	GetJobLogStream(job *brigade.Job) (io.ReadCloser, error)
	// GetWorkerLog retrieves all logs for a worker from storage.
	GetWorkerLog(job *brigade.Worker) (string, error)
	// GetWorkerLogStream retrieve a stream of all logs for a worker from storage.
	GetWorkerLogStream(job *brigade.Worker) (io.ReadCloser, error)
}
