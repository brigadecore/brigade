package controller

import (
	"io"
	"time"
)

// State is the state a build, project or job could be.
type State int

const (
	// UnknownState is in unknown state.
	UnknownState State = iota
	// PendingState is pending, didn't start.
	PendingState
	// RunningState has started and is running.
	RunningState
	// FailedState has finished and has failed.
	FailedState
	// SuccessedState has finished and has completed successfuly.
	SuccessedState
)

// Project represents a Brigade project.
type Project struct {
	Name       string
	ID         string
	LastBuilds []*Build
}

// ProjectListPageContext has the required information to
// render a project list page.
type ProjectListPageContext struct {
	Projects []*Project
	Error    error
}

// Build is a project build.
type Build struct {
	ID        string
	Version   string
	State     State
	EventType string
	Started   time.Time
	Ended     time.Time
}

// ProjectBuildListPageContext has the required information to
// render a project build list page.
type ProjectBuildListPageContext struct {
	ProjectName string
	ProjectURL  string
	ProjectNS   string

	Builds []*Build
	Error  error
}

// Job is a build job.
type Job struct {
	ID      string
	Name    string
	Image   string
	State   State
	Started time.Time
	Ended   time.Time
}

// BuildJobListPageContext has the required information to
// render a build job list page.
type BuildJobListPageContext struct {
	BuildInfo *Build
	Jobs      []*Job
	Error     error
}

// JobLogPageContext has the required information to
// render a job log page.
type JobLogPageContext struct {
	Job   *Job
	Log   io.ReadCloser
	Error error
}
