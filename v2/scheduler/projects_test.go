package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	coreTesting "github.com/brigadecore/brigade/sdk/v3/testing"
	"github.com/stretchr/testify/require"
)

func TestManageProjects(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func() *scheduler
		assertions func(error)
	}{
		{
			name: "error listing projects",
			setup: func() *scheduler {
				return &scheduler{
					config: schedulerConfig{
						addAndRemoveProjectsInterval: time.Second,
					},
					projectsClient: &coreTesting.MockProjectsClient{
						ListFn: func(
							context.Context,
							*sdk.ProjectsSelector,
							*meta.ListOptions,
						) (sdk.ProjectList, error) {
							return sdk.ProjectList{}, errors.New("something went wrong")
						},
					},
				}
			},
			assertions: func(err error) {
				require.Error(t, err)
				require.Contains(t, err.Error(), "something went wrong")
				require.Contains(t, err.Error(), "error listing projects")
			},
		},
		{
			name: "success",
			setup: func() *scheduler {
				return &scheduler{
					config: schedulerConfig{
						addAndRemoveProjectsInterval: time.Second,
					},
					projectsClient: &coreTesting.MockProjectsClient{
						ListFn: func(
							context.Context,
							*sdk.ProjectsSelector,
							*meta.ListOptions,
						) (sdk.ProjectList, error) {
							return sdk.ProjectList{
								Items: []sdk.Project{
									{
										ObjectMeta: meta.ObjectMeta{
											ID: "blue-book",
										},
									},
								},
							}, nil
						},
					},
					runWorkerLoopFn: func(context.Context, string) {},
					runJobLoopFn:    func(context.Context, string) {},
				}
			},
			assertions: func(err error) {
				require.NoError(t, err)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			scheduler := testCase.setup()
			scheduler.errCh = make(chan error)
			go scheduler.manageProjects(ctx)
			// Listen for errors
			select {
			case err := <-scheduler.errCh:
				testCase.assertions(err)
			case <-ctx.Done():
				testCase.assertions(nil)
			}
			cancel()
		})
	}
}
