package storage

import (
	"io"

	"github.com/deis/acid/pkg/acid"
)

// Store represents a storage engine for a Project.
type Store interface {
	// GetProjects retrieves all projects from storage.
	GetProjects() ([]*acid.Project, error)
	// GetProject retrieves the project from storage.
	GetProject(id string) (*acid.Project, error)
	// GetProjectBuilds retrieves the project's builds from storage.
	GetProjectBuilds(proj *acid.Project) ([]*acid.Build, error)
	// GetBuild retrieves the build from storage.
	GetBuild(id string) (*acid.Build, error)
	// CreateBuild creates a new job for the work queue.
	CreateBuild(build *acid.Build) error
	// GetBuildJobs retrieves all build jobs (pods) from storage.
	GetBuildJobs(build *acid.Build) ([]*acid.Job, error)
	// GetJob retrieves the job from storage.
	GetJob(id string) (*acid.Job, error)
	// GetJobLog retrieves all logs for a job from storage.
	GetJobLog(job *acid.Job) (string, error)
	// GetJobLogStream retrieve a stream of all logs for a job from storage.
	GetJobLogStream(job *acid.Job) (io.ReadCloser, error)
}
