package core

import (
	"context"
	"errors"
	"testing"
	"time"

	libAuthz "github.com/brigadecore/brigade/v2/apiserver/internal/lib/authz"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

func TestNewjobsService(t *testing.T) {
	projectsStore := &mockProjectsStore{}
	eventsStore := &mockEventsStore{}
	jobsStore := &mockJobsStore{}
	substrate := &mockSubstrate{}
	svc := NewJobsService(
		libAuthz.AlwaysAuthorize,
		projectsStore,
		eventsStore,
		jobsStore,
		substrate,
	)
	require.NotNil(t, svc.(*jobsService).authorize)
	require.Same(t, projectsStore, svc.(*jobsService).projectsStore)
	require.Same(t, eventsStore, svc.(*jobsService).eventsStore)
	require.Same(t, jobsStore, svc.(*jobsService).jobsStore)
	require.Same(t, substrate, svc.(*jobsService).substrate)
}

func TestJobsServiceCreate(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	testEnvironment := map[string]string{
		"FOO": "bar",
		"BAT": "baz",
	}
	testCases := []struct {
		name       string
		service    JobsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &jobsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving event from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "job with name already exists",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
									},
								},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrConflict{}, err)
			},
		},
		{
			name: "privileged container requested but not allowed",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Spec: WorkerSpec{},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
				require.Equal(
					t,
					"Worker configuration forbids jobs from utilizing privileged "+
						"containers.",
					err.(*meta.ErrAuthorization).Reason,
				)
			},
		},
		{
			name: "host docker socket mount requested but not allowed",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Spec: WorkerSpec{
									JobPolicies: &JobPolicies{
										AllowPrivileged: true,
									},
								},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
				require.Equal(
					t,
					"Worker configuration forbids jobs from mounting the Docker socket.",
					err.(*meta.ErrAuthorization).Reason,
				)
			},
		},
		{
			name: "uses workspace but worker does not",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Spec: WorkerSpec{
									JobPolicies: &JobPolicies{
										AllowPrivileged:        true,
										AllowDockerSocketMount: true,
									},
								},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrConflict{}, err)
				require.Equal(
					t,
					"The job requested access to the shared workspace, but Worker "+
						"configuration has not enabled this feature.",
					err.(*meta.ErrConflict).Reason,
				)
			},
		},
		{
			name: "error getting project from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Spec: WorkerSpec{
									UseWorkspace: true,
									JobPolicies: &JobPolicies{
										AllowPrivileged:        true,
										AllowDockerSocketMount: true,
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error creating job in store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Spec: WorkerSpec{
									UseWorkspace: true,
									JobPolicies: &JobPolicies{
										AllowPrivileged:        true,
										AllowDockerSocketMount: true,
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				jobsStore: &mockJobsStore{
					CreateFn: func(context.Context, string, Job) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error saving event")
			},
		},
		{
			name: "error storing job environment in substrate",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Spec: WorkerSpec{
									UseWorkspace: true,
									JobPolicies: &JobPolicies{
										AllowPrivileged:        true,
										AllowDockerSocketMount: true,
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				jobsStore: &mockJobsStore{
					CreateFn: func(context.Context, string, Job) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					StoreJobEnvironmentFn: func(
						context.Context,
						Project,
						string,
						string,
						JobSpec,
					) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error storing event")
			},
		},
		{
			name: "error scheduling job on substrate",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Spec: WorkerSpec{
									UseWorkspace: true,
									JobPolicies: &JobPolicies{
										AllowPrivileged:        true,
										AllowDockerSocketMount: true,
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				jobsStore: &mockJobsStore{
					CreateFn: func(context.Context, string, Job) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					StoreJobEnvironmentFn: func(
						context.Context,
						Project,
						string,
						string,
						JobSpec,
					) error {
						return nil
					},
					ScheduleJobFn: func(context.Context, Project, Event, string) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error scheduling event")
			},
		},
		{
			name: "success",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Spec: WorkerSpec{
									UseWorkspace: true,
									JobPolicies: &JobPolicies{
										AllowPrivileged:        true,
										AllowDockerSocketMount: true,
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				jobsStore: &mockJobsStore{
					CreateFn: func(_ context.Context, _ string, job Job) error {
						// Assert that all expected environment redactions occurred
						for k := range testEnvironment {
							v, ok := job.Spec.PrimaryContainer.Environment[k]
							require.True(t, ok)
							require.Equal(t, "*** REDACTED ***", v)
						}
						for _, sidecar := range job.Spec.SidecarContainers {
							for k := range testEnvironment {
								v, ok := sidecar.Environment[k]
								require.True(t, ok)
								require.Equal(t, "*** REDACTED ***", v)
							}
						}
						return nil
					},
				},
				substrate: &mockSubstrate{
					StoreJobEnvironmentFn: func(
						_ context.Context,
						_ Project,
						_ string,
						_ string,
						jobSpec JobSpec,
					) error {
						// Assert that an object WITHOUT environment redactions was received
						require.Equal(
							t,
							testEnvironment,
							jobSpec.PrimaryContainer.Environment,
						)
						for _, sidecar := range jobSpec.SidecarContainers {
							require.Equal(t, testEnvironment, sidecar.Environment)
						}
						return nil
					},
					ScheduleJobFn: func(context.Context, Project, Event, string) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(
				testCase.service.Create(
					context.Background(),
					testEventID,
					Job{
						Name: testJobName,
						Spec: JobSpec{
							PrimaryContainer: JobContainerSpec{
								ContainerSpec: ContainerSpec{
									Environment: testEnvironment,
								},
								Privileged:          true,
								UseHostDockerSocket: true,
								WorkspaceMountPath:  "/var/workspace",
							},
							SidecarContainers: map[string]JobContainerSpec{
								"foo": {
									ContainerSpec: ContainerSpec{
										Environment: testEnvironment,
									},
									Privileged:          true,
									UseHostDockerSocket: true,
									WorkspaceMountPath:  "/var/workspace",
								},
							},
						},
					},
				),
			)
		})
	}
}

func TestJobsServiceCreateRetry(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	testCases := []struct {
		name               string
		service            JobsService
		workspaceMountPath string
		assertions         func(error)
	}{
		{
			name: "job retry - not equivalent",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Labels: map[string]string{
								RetryLabelKey: testEventID,
							},
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Spec: JobSpec{
											PrimaryContainer: JobContainerSpec{
												WorkspaceMountPath: "",
											},
											SidecarContainers: map[string]JobContainerSpec{
												// The original job names this "foo"
												"bar": {
													WorkspaceMountPath: "",
												},
											},
										},
										Status: &JobStatus{
											Phase: JobPhaseSucceeded,
										},
									},
								},
							},
						}, nil
					},
				},
			},
			workspaceMountPath: "",
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrConflict{}, err)
			},
		},
		{
			name: "job retry success - no-op",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Labels: map[string]string{
								RetryLabelKey: testEventID,
							},
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Spec: JobSpec{
											PrimaryContainer: JobContainerSpec{
												WorkspaceMountPath: "",
											},
											SidecarContainers: map[string]JobContainerSpec{
												"foo": {
													WorkspaceMountPath: "",
												},
											},
										},
										Status: &JobStatus{
											Phase: JobPhaseSucceeded,
										},
									},
								},
							},
						}, nil
					},
				},
				// No other methods mocked out; they should not be called
			},
			workspaceMountPath: "",
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(
				testCase.service.Create(
					context.Background(),
					testEventID,
					Job{
						Name: testJobName,
						Spec: JobSpec{
							PrimaryContainer: JobContainerSpec{
								WorkspaceMountPath: testCase.workspaceMountPath,
							},
							SidecarContainers: map[string]JobContainerSpec{
								"foo": {
									WorkspaceMountPath: testCase.workspaceMountPath,
								},
							},
						},
						Status: &JobStatus{
							Phase: JobPhaseSucceeded,
						},
					},
				),
			)
		})
	}
}

func TestJobsServiceStart(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "foo"
	testCases := []struct {
		name       string
		service    JobsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &jobsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "event has no such job",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
			},
		},
		{
			name: "job is not currently pending",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Status: &JobStatus{
											Phase: JobPhaseRunning,
										},
									},
								},
							},
						}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrConflict{}, err)
			},
		},
		{
			name: "error getting project from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Status: &JobStatus{
											Phase: JobPhasePending,
										},
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error updating job status",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Status: &JobStatus{
											Phase: JobPhasePending,
										},
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						context.Context,
						string, string,
						JobStatus,
					) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating status of event")
			},
		},
		{
			name: "error starting job on substrate",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Status: &JobStatus{
											Phase: JobPhasePending,
										},
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						context.Context,
						string, string,
						JobStatus,
					) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					StartJobFn: func(context.Context, Project, Event, string) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error starting event")
			},
		},
		{
			name: "success",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Status: &JobStatus{
											Phase: JobPhasePending,
										},
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						context.Context,
						string, string,
						JobStatus,
					) error {
						return nil
					},
				},
				substrate: &mockSubstrate{
					StartJobFn: func(context.Context, Project, Event, string) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err :=
				testCase.service.Start(context.Background(), testEventID, testJobName)
			testCase.assertions(err)
		})
	}
}

func TestJobsServiceGetStatus(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	testJobStatus := JobStatus{
		Phase: JobPhaseRunning,
	}
	testCases := []struct {
		name       string
		service    JobsService
		assertions func(JobStatus, error)
	}{
		{
			name: "unauthorized",
			service: &jobsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(_ JobStatus, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ JobStatus, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "job not found",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(_ JobStatus, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
			},
		},
		{
			name: "success",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name:   testJobName,
										Status: &testJobStatus,
									},
								},
							},
						}, nil
					},
				},
			},
			assertions: func(status JobStatus, err error) {
				require.NoError(t, err)
				require.Equal(t, testJobStatus, status)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.assertions(
				testCase.service.GetStatus(
					context.Background(),
					testEventID,
					testJobName,
				),
			)
		})
	}
}

func TestJobsServiceWatchStatus(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	testJobStatus := JobStatus{
		Phase: JobPhaseRunning,
	}
	testCases := []struct {
		name       string
		service    JobsService
		assertions func(context.Context, <-chan JobStatus, error)
	}{
		{
			name: "unauthorized",
			service: &jobsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(_ context.Context, _ <-chan JobStatus, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(_ context.Context, _ <-chan JobStatus, err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "job not found",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(_ context.Context, _ <-chan JobStatus, err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
			},
		},
		{
			name: "success",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name:   testJobName,
										Status: &testJobStatus,
									},
								},
							},
						}, nil
					},
				},
			},
			assertions: func(
				ctx context.Context,
				statusCh <-chan JobStatus,
				err error,
			) {
				require.NoError(t, err)
				select {
				case status := <-statusCh:
					require.Equal(t, testJobStatus, status)
				case <-ctx.Done():
					require.Fail(t, "didn't receive status update over channel")
				}
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			statusCh, err :=
				testCase.service.WatchStatus(ctx, testEventID, testJobName)
			testCase.assertions(ctx, statusCh, err)
			cancel()
		})
	}
}

func TestJobsServiceUpdateStatus(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	testCases := []struct {
		name       string
		service    JobsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &jobsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error retrieving event from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "job not found",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), `Job "italian" not found`)
			},
		},
		{
			name: "error updating job in store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Status: &JobStatus{
											Phase: JobPhasePending,
										},
									},
								},
							},
						}, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						context.Context,
						string,
						string,
						JobStatus,
					) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating status of event")
			},
		},
		{
			name: "job's phase already terminal",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Status: &JobStatus{
											Phase: JobPhaseCanceled,
										},
									},
								},
							},
						}, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						context.Context,
						string,
						string,
						JobStatus,
					) error {
						require.Fail(
							t,
							"UpdateStatusFn should not have been called, but was",
						)
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(
					t,
					err.Error(),
					`job "italian" has already reached a terminal phase`,
				)
			},
		},
		{
			name: "success",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
										Status: &JobStatus{
											Phase: JobPhaseRunning,
										},
									},
								},
							},
						}, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						context.Context,
						string,
						string,
						JobStatus,
					) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.UpdateStatus(
				context.Background(),
				testEventID,
				testJobName,
				JobStatus{},
			)
			testCase.assertions(err)
		})
	}
}

func TestJobsServiceCleanup(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	testCases := []struct {
		name       string
		service    JobsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &jobsService{
				authorize: libAuthz.NeverAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error getting event from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving event")
			},
		},
		{
			name: "event has no such job",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{}, nil
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrNotFound{}, err)
			},
		},
		{
			name: "error getting project from store",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error retrieving project")
			},
		},
		{
			name: "error deleting job from substrate",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					DeleteJobFn: func(context.Context, Project, Event, string) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error deleting event")
			},
		},
		{
			name: "success",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return Event{
							Worker: Worker{
								Jobs: []Job{
									{
										Name: testJobName,
									},
								},
							},
						}, nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					DeleteJobFn: func(context.Context, Project, Event, string) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.Cleanup(
				context.Background(),
				testEventID,
				testJobName,
			)
			testCase.assertions(err)
		})
	}
}

func TestJobsServiceTimeout(t *testing.T) {
	const testEventID = "123456789"
	const testJobName = "italian"
	var testStartedTime = time.Unix(1234, 56789)
	var testEvent = Event{
		Worker: Worker{
			Jobs: []Job{
				{
					Name: testJobName,
					Status: &JobStatus{
						Started: &testStartedTime,
						Phase:   JobPhaseRunning,
					},
				},
			},
		},
	}
	testCases := []struct {
		name       string
		service    JobsService
		assertions func(error)
	}{
		{
			name: "unauthorized",
			service: &jobsService{
				authorize: libAuthz.NeverAuthorize,
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name: "error updating job status",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return testEvent, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						context.Context,
						string,
						string,
						JobStatus,
					) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error updating status")
			},
		},
		{
			name: "error cleaning up",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return testEvent, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						context.Context,
						string,
						string,
						JobStatus,
					) error {
						return nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					DeleteJobFn: func(context.Context, Project, Event, string) error {
						return errors.New("something went wrong")
					},
				},
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error deleting event")
			},
		},
		{
			name: "success",
			service: &jobsService{
				authorize: libAuthz.AlwaysAuthorize,
				eventsStore: &mockEventsStore{
					GetFn: func(context.Context, string) (Event, error) {
						return testEvent, nil
					},
				},
				jobsStore: &mockJobsStore{
					UpdateStatusFn: func(
						_ context.Context,
						_ string,
						_ string,
						status JobStatus,
					) error {
						require.Equal(t, JobPhaseTimedOut, status.Phase)
						require.Equal(t, &testStartedTime, status.Started)
						require.NotNil(t, status.Ended)
						return nil
					},
				},
				projectsStore: &mockProjectsStore{
					GetFn: func(context.Context, string) (Project, error) {
						return Project{}, nil
					},
				},
				substrate: &mockSubstrate{
					DeleteJobFn: func(context.Context, Project, Event, string) error {
						return nil
					},
				},
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.service.Timeout(
				context.Background(),
				testEventID,
				testJobName,
			)
			testCase.assertions(err)
		})
	}
}

type mockJobsStore struct {
	CreateFn       func(ctx context.Context, eventID string, job Job) error
	UpdateStatusFn func(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
}

func (m *mockJobsStore) Create(
	ctx context.Context,
	eventID string,
	job Job,
) error {
	return m.CreateFn(ctx, eventID, job)
}

func (m *mockJobsStore) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status JobStatus,
) error {
	return m.UpdateStatusFn(ctx, eventID, jobName, status)
}
