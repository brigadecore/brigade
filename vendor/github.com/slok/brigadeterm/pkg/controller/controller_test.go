package controller_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"
	"time"

	azurebrigade "github.com/brigadecore/brigade/pkg/brigade"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/slok/brigadeterm/pkg/controller"
	mbrigade "github.com/slok/brigadeterm/pkg/mocks/service/brigade"
	brigademodel "github.com/slok/brigadeterm/pkg/model/brigade"
)

func TestControllerProjectListPageContext(t *testing.T) {
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now().Add(-4 * time.Minute)

	tests := []struct {
		name       string
		projects   []*brigademodel.Project
		lastBuilds []*brigademodel.Build
		expCtx     *controller.ProjectListPageContext
	}{
		{
			name: "One project should return one project context.",
			projects: []*brigademodel.Project{
				&brigademodel.Project{
					ID:   "prj1",
					Name: "project-1",
				},
			},
			lastBuilds: []*brigademodel.Build{
				{
					ID:       "build1",
					Revision: &azurebrigade.Revision{Commit: "1234567890"},
					Worker: &azurebrigade.Worker{
						Status:    azurebrigade.JobSucceeded,
						StartTime: start,
						EndTime:   end,
					},
					Type: "testEvent",
				},
				{
					ID:       "build2",
					Revision: &azurebrigade.Revision{Commit: "1234567890"},
					Worker: &azurebrigade.Worker{
						Status:    azurebrigade.JobFailed,
						StartTime: start,
						EndTime:   end,
					},
					Type: "testEvent",
				},
				{
					ID:       "build3",
					Revision: &azurebrigade.Revision{Commit: "1234567890"},
					Worker: &azurebrigade.Worker{
						Status:    azurebrigade.JobRunning,
						StartTime: start,
						EndTime:   end,
					},
					Type: "testEvent",
				},
			},
			expCtx: &controller.ProjectListPageContext{
				Projects: []*controller.Project{
					&controller.Project{
						ID:   "prj1",
						Name: "project-1",
						LastBuilds: []*controller.Build{
							{
								ID:        "build1",
								Version:   "1234567890",
								EventType: "testEvent",
								State:     controller.SuccessedState,
								Started:   start,
								Ended:     end,
							},
							{
								ID:        "build2",
								Version:   "1234567890",
								EventType: "testEvent",
								State:     controller.FailedState,
								Started:   start,
								Ended:     end,
							},
							{
								ID:        "build3",
								Version:   "1234567890",
								EventType: "testEvent",
								State:     controller.RunningState,
								Started:   start,
								Ended:     end,
							},
						},
					},
				},
			},
		},
		{
			name: "Multiple projects should return multiple project context.",
			projects: []*brigademodel.Project{
				&brigademodel.Project{
					ID:   "prj1",
					Name: "project-1",
				},
				&brigademodel.Project{
					ID:   "prj2",
					Name: "project-2",
				},
			},
			lastBuilds: []*brigademodel.Build{
				{
					ID:       "build1",
					Revision: &azurebrigade.Revision{Commit: "1234567890"},
					Worker: &azurebrigade.Worker{
						Status:    azurebrigade.JobFailed,
						StartTime: start,
						EndTime:   end,
					},
					Type: "testEvent",
				},
			},
			expCtx: &controller.ProjectListPageContext{
				Projects: []*controller.Project{
					&controller.Project{
						ID:   "prj1",
						Name: "project-1",
						LastBuilds: []*controller.Build{
							{
								ID:        "build1",
								Version:   "1234567890",
								EventType: "testEvent",
								State:     controller.FailedState,
								Started:   start,
								Ended:     end,
							},
						},
					},
					&controller.Project{
						ID:   "prj2",
						Name: "project-2",
						LastBuilds: []*controller.Build{
							{
								ID:        "build1",
								Version:   "1234567890",
								EventType: "testEvent",
								State:     controller.FailedState,
								Started:   start,
								Ended:     end,
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			mb := &mbrigade.Service{}
			mb.On("GetProjects").Return(test.projects, nil)
			mb.On("GetProjectLastBuilds", mock.Anything, mock.Anything).Return(test.lastBuilds, nil)

			c := controller.NewController(mb)
			ctx := c.ProjectListPageContext()

			assert.Equal(test.expCtx, ctx)
		})
	}
}

func TestControllerProjectBuildListPageContext(t *testing.T) {
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now().Add(-4 * time.Minute)

	tests := []struct {
		name    string
		project *brigademodel.Project
		builds  []*brigademodel.Build
		expCtx  *controller.ProjectBuildListPageContext
	}{
		{
			name: "Multiple builds should return multiple builds context.",
			project: &brigademodel.Project{
				ID:         "prj1",
				Name:       "project-1",
				Kubernetes: azurebrigade.Kubernetes{Namespace: "test"},
				Repo:       azurebrigade.Repo{CloneURL: "git@github.com:slok/brigadeterm"},
			},
			builds: []*brigademodel.Build{
				&brigademodel.Build{
					ID:       "build1",
					Revision: &azurebrigade.Revision{Commit: "1234567890"},
					Worker: &azurebrigade.Worker{
						Status:    azurebrigade.JobFailed,
						StartTime: start,
						EndTime:   end,
					},
					Type: "testEvent",
				},
				&brigademodel.Build{
					ID:       "build2",
					Revision: &azurebrigade.Revision{Commit: "1234567890"},
					Worker: &azurebrigade.Worker{
						Status:    azurebrigade.JobSucceeded,
						StartTime: start,
						EndTime:   end,
					},
					Type: "testEvent",
				},
				&brigademodel.Build{
					ID:       "build3",
					Revision: &azurebrigade.Revision{Commit: "1234567890"},
					Worker: &azurebrigade.Worker{
						Status:    azurebrigade.JobRunning,
						StartTime: start,
					},
					Type: "testEvent",
				},
			},
			expCtx: &controller.ProjectBuildListPageContext{
				ProjectName: "project-1",
				ProjectNS:   "test",
				ProjectURL:  "git@github.com:slok/brigadeterm",
				Builds: []*controller.Build{
					&controller.Build{
						ID:        "build1",
						Version:   "1234567890",
						State:     controller.FailedState,
						EventType: "testEvent",
						Started:   start,
						Ended:     end,
					},
					&controller.Build{
						ID:        "build2",
						Version:   "1234567890",
						State:     controller.SuccessedState,
						EventType: "testEvent",
						Started:   start,
						Ended:     end,
					},
					&controller.Build{
						ID:        "build3",
						Version:   "1234567890",
						State:     controller.RunningState,
						EventType: "testEvent",
						Started:   start,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			mb := &mbrigade.Service{}
			mb.On("GetProject", mock.Anything).Return(test.project, nil)
			mb.On("GetProjectBuilds", mock.Anything, mock.Anything).Return(test.builds, nil)

			c := controller.NewController(mb)
			ctx := c.ProjectBuildListPageContext("whatever")

			assert.Equal(test.expCtx, ctx)
		})
	}
}

func TestControllerBuildJobListPageContext(t *testing.T) {
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now().Add(-4 * time.Minute)

	tests := []struct {
		name   string
		build  *brigademodel.Build
		jobs   []*brigademodel.Job
		expCtx *controller.BuildJobListPageContext
	}{
		{
			name: "Multiple jobs should return multiple jobs context.",
			build: &brigademodel.Build{
				ID:       "build1",
				Revision: &azurebrigade.Revision{Commit: "1234567890"},
				Worker: &azurebrigade.Worker{
					Status:    azurebrigade.JobSucceeded,
					StartTime: start,
					EndTime:   end,
				},
				Type: "testEvent",
			},
			jobs: []*brigademodel.Job{
				&brigademodel.Job{
					ID:        "j1",
					Name:      "job-1",
					Image:     "myimage/image:v0.1.0",
					Status:    azurebrigade.JobRunning,
					StartTime: start,
					EndTime:   end,
				},
				&brigademodel.Job{
					ID:        "j2",
					Name:      "job-2",
					Image:     "myimage/image:v0.2.0",
					Status:    azurebrigade.JobPending,
					StartTime: start,
					EndTime:   end,
				},
				&brigademodel.Job{
					ID:        "j3",
					Name:      "job-3",
					Image:     "myimage/image:v0.3.0",
					Status:    azurebrigade.JobSucceeded,
					StartTime: start,
					EndTime:   end,
				},
				&brigademodel.Job{
					ID:        "j4",
					Name:      "job-4",
					Image:     "myimage/image:v0.4.0",
					Status:    azurebrigade.JobFailed,
					StartTime: start,
					EndTime:   end,
				},
			},
			expCtx: &controller.BuildJobListPageContext{
				BuildInfo: &controller.Build{
					ID:        "build1",
					Version:   "1234567890",
					State:     controller.SuccessedState,
					EventType: "testEvent",
					Started:   start,
					Ended:     end,
				},
				Jobs: []*controller.Job{
					&controller.Job{
						ID:      "j1",
						Name:    "job-1",
						Image:   "myimage/image:v0.1.0",
						State:   controller.RunningState,
						Started: start,
						Ended:   end,
					},
					&controller.Job{
						ID:      "j2",
						Name:    "job-2",
						Image:   "myimage/image:v0.2.0",
						State:   controller.PendingState,
						Started: start,
						Ended:   end,
					},
					&controller.Job{
						ID:      "j3",
						Name:    "job-3",
						Image:   "myimage/image:v0.3.0",
						State:   controller.SuccessedState,
						Started: start,
						Ended:   end,
					},
					&controller.Job{
						ID:      "j4",
						Name:    "job-4",
						Image:   "myimage/image:v0.4.0",
						State:   controller.FailedState,
						Started: start,
						Ended:   end,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			mb := &mbrigade.Service{}
			mb.On("GetBuild", mock.Anything).Return(test.build, nil)
			mb.On("GetBuildJobs", mock.Anything, mock.Anything).Return(test.jobs, nil)

			c := controller.NewController(mb)
			ctx := c.BuildJobListPageContext("whatever")

			assert.Equal(test.expCtx, ctx)
		})
	}
}

func TestControllerJobLogPageContext(t *testing.T) {
	start := time.Now().Add(-5 * time.Minute)
	end := time.Now().Add(-4 * time.Minute)

	tests := []struct {
		name         string
		job          *brigademodel.Job
		log          io.ReadCloser
		expCtx       *controller.JobLogPageContext
		expStreaming bool
	}{
		{
			name: "job log should return job's log context with a stream when running.",
			job: &brigademodel.Job{
				ID:        "j1",
				Name:      "job-1",
				Image:     "myimage/image:v0.1.0",
				Status:    azurebrigade.JobRunning,
				StartTime: start,
				EndTime:   end,
			},
			log:          ioutil.NopCloser(bytes.NewBufferString("my awesome log")),
			expStreaming: true,
			expCtx: &controller.JobLogPageContext{
				Job: &controller.Job{
					ID:      "j1",
					Name:    "job-1",
					Image:   "myimage/image:v0.1.0",
					State:   controller.RunningState,
					Started: start,
					Ended:   end,
				},
				Log: ioutil.NopCloser(bytes.NewBufferString("my awesome log")),
			},
		},
		{
			name: "job log should return job's log context without a stream when not complete.",
			job: &brigademodel.Job{
				ID:        "j1",
				Name:      "job-1",
				Image:     "myimage/image:v0.1.0",
				Status:    azurebrigade.JobSucceeded,
				StartTime: start,
				EndTime:   end,
			},
			log:          ioutil.NopCloser(bytes.NewBufferString("my awesome log")),
			expStreaming: false,
			expCtx: &controller.JobLogPageContext{
				Job: &controller.Job{
					ID:      "j1",
					Name:    "job-1",
					Image:   "myimage/image:v0.1.0",
					State:   controller.SuccessedState,
					Started: start,
					Ended:   end,
				},
				Log: ioutil.NopCloser(bytes.NewBufferString("my awesome log")),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			mb := &mbrigade.Service{}
			mb.On("GetJob", mock.Anything).Return(test.job, nil)
			if test.expStreaming {
				mb.On("GetJobLogStream", mock.Anything).Return(test.log, nil)
			} else {
				mb.On("GetJobLog", mock.Anything).Return(test.log, nil)
			}

			c := controller.NewController(mb)
			ctx := c.JobLogPageContext("whatever")

			assert.Equal(test.expCtx, ctx)
		})
	}
}

func TestControllerJobRunning(t *testing.T) {
	tests := []struct {
		name string
		job  *brigademodel.Job
		exp  bool
	}{
		{
			name: "Job running should return true",
			job: &brigademodel.Job{
				Status: azurebrigade.JobRunning,
			},
			exp: true,
		},
		{
			name: "Job unknown should return false",
			job: &brigademodel.Job{
				Status: azurebrigade.JobUnknown,
			},
			exp: false,
		},
		{
			name: "Job failed should return false",
			job: &brigademodel.Job{
				Status: azurebrigade.JobFailed,
			},
			exp: false,
		},
		{
			name: "Job Succeeded should return false",
			job: &brigademodel.Job{
				Status: azurebrigade.JobSucceeded,
			},
			exp: false,
		},
		{
			name: "Job Pending should return false",
			job: &brigademodel.Job{
				Status: azurebrigade.JobPending,
			},
			exp: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			// Mocks.
			mb := &mbrigade.Service{}
			mb.On("GetJob", mock.Anything).Return(test.job, nil)

			c := controller.NewController(mb)
			got := c.JobRunning("whatever")
			assert.Equal(test.exp, got)
		})
	}
}
