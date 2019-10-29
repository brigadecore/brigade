package controller

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	azurebrigade "github.com/brigadecore/brigade/pkg/brigade"

	brigademodel "github.com/slok/brigadeterm/pkg/model/brigade"
	"github.com/slok/brigadeterm/pkg/service/brigade"
)

const (
	projectLastBuildsQuantity = 5
)

// Controller knows what to how to handle the different ui views data
// using the required services and having the logic of each part.
type Controller interface {
	// ProjectListPageContext returns the projectListPage context.
	ProjectListPageContext() *ProjectListPageContext
	// ProjectBuildListContext returns the projectBuildListPage context.
	ProjectBuildListPageContext(projectID string) *ProjectBuildListPageContext
	// BuildJobListPageContext returns the BuildJobListPage context.
	BuildJobListPageContext(buildID string) *BuildJobListPageContext
	// JobLogPageContext returns the JobLogPage context.
	JobLogPageContext(jobID string) *JobLogPageContext
	// JobRunning returns if the job is running or finished.
	JobRunning(jobID string) bool
	// RerunBuild will create a new build based on the build ID.
	RerunBuild(buildID string) error
}

type controller struct {
	brigade brigade.Service
}

// NewController returns a new controller.
func NewController(brigade brigade.Service) Controller {
	return &controller{
		brigade: brigade,
	}
}

func (c *controller) ProjectListPageContext() *ProjectListPageContext {
	prjs, err := c.brigade.GetProjects()
	if err != nil {
		return &ProjectListPageContext{
			Error: fmt.Errorf("there was an error while getting projects from brigade: %s", err),
		}
	}

	ctxPrjs := make([]*Project, len(prjs))
	for i, prj := range prjs {
		p := &Project{
			ID:   prj.ID,
			Name: prj.Name,
		}
		ctxPrjs[i] = p

		// Set last build of the project.
		lastBuilds, err := c.brigade.GetProjectLastBuilds(prj.ID, projectLastBuildsQuantity)
		if err != nil {
			continue
		}

		lb := []*Build{}
		for _, b := range lastBuilds {
			lb = append(lb, c.transformBuild(b))
		}
		p.LastBuilds = lb
	}

	return &ProjectListPageContext{
		Projects: ctxPrjs,
	}
}

func (c *controller) ProjectBuildListPageContext(projectID string) *ProjectBuildListPageContext {
	prj, err := c.brigade.GetProject(projectID)
	if err != nil {
		return &ProjectBuildListPageContext{
			Error: fmt.Errorf("there was an error while getting project from brigade: %s", err),
		}
	}

	builds, err := c.brigade.GetProjectBuilds(prj, true)
	if err != nil {
		return &ProjectBuildListPageContext{
			Error: fmt.Errorf("there was an error while getting builds from brigade: %s", err),
		}
	}

	ctxBuilds := make([]*Build, len(builds))
	for i, b := range builds {
		ctxBuilds[i] = c.transformBuild(b)
	}

	return &ProjectBuildListPageContext{
		ProjectName: prj.Name,
		ProjectNS:   prj.Kubernetes.Namespace,
		ProjectURL:  prj.Repo.CloneURL,
		Builds:      ctxBuilds,
	}
}

func (c *controller) BuildJobListPageContext(buildID string) *BuildJobListPageContext {
	build, err := c.brigade.GetBuild(buildID)

	jobs, err := c.brigade.GetBuildJobs(buildID, false)
	if err != nil {
		return &BuildJobListPageContext{
			Error: fmt.Errorf("there was an error while getting the jobs from brigade: %s", err),
		}
	}

	ctxBuild := c.transformBuild(build)

	ctxJobs := make([]*Job, len(jobs))
	for i, j := range jobs {
		ctxJobs[i] = c.transformJob(j)
	}

	return &BuildJobListPageContext{
		BuildInfo: ctxBuild,
		Jobs:      ctxJobs,
	}
}

func (c *controller) JobLogPageContext(jobID string) *JobLogPageContext {
	job, err := c.brigade.GetJob(jobID)
	if err != nil {
		return &JobLogPageContext{
			Error: fmt.Errorf("there was an error while getting %s job: %s", jobID, err),
		}
	}

	var logStrm io.ReadCloser

	// If the job is running then get a stream of logs.
	if c.transformState(job.Status) == RunningState {
		logStrm, err = c.brigade.GetJobLogStream(jobID)
	} else {
		logStrm, err = c.brigade.GetJobLog(jobID)
	}

	if err != nil {
		return &JobLogPageContext{
			Error: fmt.Errorf("there was an error while getting %s job log: %s", jobID, err),
		}
	}

	// Set a dummy ReadCloser for safety reads on the variable.
	if logStrm == nil {
		logStrm = ioutil.NopCloser(new(bytes.Buffer))
	}

	return &JobLogPageContext{
		Job: c.transformJob(job),
		Log: logStrm,
	}
}

func (c *controller) JobRunning(jobID string) bool {
	job, err := c.brigade.GetJob(jobID)
	// If error assume the job is not running or doesn't exist so it's finished.
	if err != nil {
		return false
	}

	if c.transformState(job.Status) == RunningState {
		return true
	}

	return false
}

func (c *controller) transformBuild(b *brigademodel.Build) *Build {
	var start time.Time
	var end time.Time
	var state = UnknownState

	if b.Worker != nil {
		start = b.Worker.StartTime
		end = b.Worker.EndTime
		state = c.transformState(b.Worker.Status)
	}

	return &Build{
		ID:        b.ID,
		Version:   b.Revision.Commit,
		State:     state,
		EventType: b.Type,
		Started:   start,
		Ended:     end,
	}
}

func (c *controller) transformJob(j *brigademodel.Job) *Job {

	return &Job{
		ID:      j.ID,
		Name:    j.Name,
		Image:   j.Image,
		State:   c.transformState(j.Status),
		Started: j.StartTime,
		Ended:   j.EndTime,
	}
}

func (c *controller) transformState(st brigademodel.State) State {
	switch st {
	case azurebrigade.JobPending:
		return PendingState
	case azurebrigade.JobRunning:
		return RunningState
	case azurebrigade.JobSucceeded:
		return SuccessedState
	case azurebrigade.JobFailed:
		return FailedState
	default:
		return UnknownState
	}
}

func (c *controller) RerunBuild(buildID string) error {
	return c.brigade.RerunBuild(buildID)
}
