// +build integration

package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/stretchr/testify/require"
)

const (
	testJobName = "test-job"
	testTimeout = time.Duration(120 * time.Second)
)

type testcase struct {
	name                       string
	project                    core.Project
	configFiles                map[string]string
	expectedJobLogsContains    string
	expectedJobPhase           core.JobPhase
	expectedWorkerLogsContains string
	expectedWorkerPhase        core.WorkerPhase
	expectedVCSLogsContains    string
}

var testcases = []testcase{
	{
		name: "GitHub - no ref",
		project: core.Project{
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
					},
				},
			},
		},
		expectedJobLogsContains: "README.md",
		expectedJobPhase:        core.JobPhaseSucceeded,
		expectedWorkerPhase:     core.WorkerPhaseSucceeded,
	},
	{
		name: "GitHub - full ref",
		project: core.Project{
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
						Ref:      "refs/heads/master",
					},
				},
			},
		},
		expectedJobLogsContains: "README.md",
		expectedJobPhase:        core.JobPhaseSucceeded,
		expectedWorkerPhase:     core.WorkerPhaseSucceeded,
	},
	{
		name: "GitHub - casual ref",
		project: core.Project{
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
						Ref:      "master",
					},
				},
			},
		},
		expectedJobLogsContains: "README.md",
		expectedJobPhase:        core.JobPhaseSucceeded,
		expectedWorkerPhase:     core.WorkerPhaseSucceeded,
	},
	{
		name: "GitHub - commit sha",
		project: core.Project{
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
						Commit:   "589e15029e1e44dee48de4800daf1f78e64287c0",
					},
				},
			},
		},
		expectedJobLogsContains: "README.md",
		expectedJobPhase:        core.JobPhaseSucceeded,
		expectedWorkerPhase:     core.WorkerPhaseSucceeded,
	},
	{
		name: "GitHub - submodules",
		project: core.Project{
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL:       "https://github.com/brigadecore/empty-testbed.git",
						InitSubmodules: true,
					},
				},
			},
		},
		expectedJobLogsContains: ".submodules",
		expectedJobPhase:        core.JobPhaseSucceeded,
		expectedWorkerPhase:     core.WorkerPhaseSucceeded,
	},
	{
		name: "GitHub - vcs failure",
		project: core.Project{
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
						Ref:      "non-existent",
					},
				},
			},
		},
		expectedWorkerPhase:     core.WorkerPhaseFailed,
		expectedVCSLogsContains: `reference "non-existent" not found in repo "https://github.com/brigadecore/empty-testbed.git"`,
	},
	{
		name: "GitHub - job fails",
		project: core.Project{
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
					},
				},
			},
		},
		configFiles: map[string]string{
			"brigade.ts": fmt.Sprintf(`
			import { events, Job } from "@brigadecore/brigadier"

			events.on("github.com/brigadecore/brigade/cli", "exec", async event => {
				let job = new Job("%s", "alpine", event)
				job.primaryContainer.command = ["sh"]
				job.primaryContainer.arguments = ["-c", "'echo Goodbye World && exit 1'"]
				await job.run()
			});
		`, testJobName)},
		expectedJobLogsContains: "Goodbye World",
		expectedJobPhase:        core.JobPhaseFailed,
		expectedWorkerPhase:     core.WorkerPhaseFailed,
	},
}

var defaultConfigFiles = map[string]string{
	"brigade.ts": fmt.Sprintf(`
		import { events, Job } from "@brigadecore/brigadier"

		events.on("github.com/brigadecore/brigade/cli", "exec", async event => {
			let job = new Job("%s", "alpine", event)
			job.primaryContainer.sourceMountPath = "/var/vcs"
			job.primaryContainer.command = ["ls"]
			job.primaryContainer.arguments = ["-haltr", "/var/vcs"]
			await job.run()
		});
	`, testJobName)}

func TestMain(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			// TODO: send in/parameterize this value
			apiServerAddress := "https://localhost:7000"
			// TODO: send in/parameterize this value
			rootPassword := "F00Bar!!!"
			apiClientOpts := &restmachinery.APIClientOptions{
				AllowInsecureConnections: true,
			}

			authClient := authn.NewSessionsClient(
				apiServerAddress,
				"",
				apiClientOpts,
			)

			token, err := authClient.CreateRootSession(ctx, rootPassword)
			require.NoError(t, err, "error creating root session")
			tokenStr := token.Value

			client := sdk.NewAPIClient(
				apiServerAddress,
				tokenStr,
				apiClientOpts,
			)

			// Update the project with defaults
			tc.project.ID = "test-project"
			if len(tc.configFiles) > 0 {
				tc.project.Spec.WorkerTemplate.DefaultConfigFiles = tc.configFiles
			} else {
				tc.project.Spec.WorkerTemplate.DefaultConfigFiles = defaultConfigFiles
			}

			// Delete the test project (we're sharing the name between tests)
			err = client.Core().Projects().Delete(ctx, tc.project.ID)
			require.NoError(t, err, "error deleting project")

			// Create the test project
			_, err = client.Core().Projects().Create(ctx, tc.project)
			require.NoError(t, err, "error creating project")

			// Verify there are no events for this project
			eList, err := client.Core().Events().List(
				ctx,
				&core.EventsSelector{ProjectID: tc.project.ID},
				&meta.ListOptions{Limit: 1},
			)
			require.Equal(t, 0, len(eList.Items), "event list items should be exactly zero")

			// Create a new event
			event := core.Event{
				ProjectID: tc.project.ID,
				Source:    "github.com/brigadecore/brigade/cli",
				Type:      "exec",
			}

			eList, err = client.Core().Events().Create(ctx, event)
			require.NoError(t, err, "error creating event")
			require.Equal(t, 1, len(eList.Items), "event list items should be exactly one")

			e := eList.Items[0]

			assertWorkerTerminalPhase(t, ctx, client, e, tc)
			assertWorkerLogs(t, ctx, client, e, tc)

			// We expect the vcs init container to fail for this test,
			// therefore, assertions on jobs are not applicable
			if tc.name == "GitHub - vcs failure" {
				return
			}

			assertJobTerminalPhase(t, ctx, client, e, tc)
			assertJobLogs(t, ctx, client, e, tc)
		})
	}
}

func assertWorkerTerminalPhase(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	tc testcase,
) {
	statusCh, errCh, err := client.Core().Events().Workers().WatchStatus(
		ctx,
		e.ID,
	)
	require.NoError(
		t,
		err,
		"error encountered attempting to watch worker status",
	)

	timer := time.NewTimer(testTimeout)
	defer timer.Stop()

	for {
		select {
		case status := <-statusCh:
			phase := status.Phase
			if phase.IsTerminal() {
				require.Equal(
					t,
					tc.expectedWorkerPhase,
					phase,
					"worker's terminal phase does not match expected",
				)
				return
			}
		case err := <-errCh:
			t.Fatalf("error encountered watching worker status: %s", err)
		case <-timer.C:
			t.Fatal("timeout waiting for worker to reach a terminal phase")
		}
	}
}

func assertWorkerLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	tc testcase,
) {
	// Check vcs container logs
	assertLogs(
		t,
		ctx,
		client,
		e,
		&core.LogsSelector{Container: "vcs"},
		tc.expectedVCSLogsContains,
	)

	// We expect the vcs init container to fail for this test,
	// therefore, assertions on worker logs are not applicable
	if tc.name == "GitHub - vcs failure" {
		return
	}

	// Check worker container logs
	var contains string
	if tc.expectedWorkerLogsContains != "" {
		contains = tc.expectedWorkerLogsContains
	} else {
		// Look for default logs that we expect must exist
		contains = "brigade-worker version"
	}
	assertLogs(
		t,
		ctx,
		client,
		e,
		&core.LogsSelector{},
		contains,
	)
}

func assertJobTerminalPhase(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	tc testcase,
) {
	statusCh, errCh, err := client.Core().Events().Workers().Jobs().WatchStatus(
		ctx,
		e.ID,
		testJobName,
	)
	require.NoError(t, err, "error encountered attempting to watch job status")

	timer := time.NewTimer(testTimeout)
	defer timer.Stop()

	for {
		select {
		case status := <-statusCh:
			phase := status.Phase
			if phase.IsTerminal() {
				require.Equal(
					t,
					tc.expectedJobPhase,
					phase,
					"job's terminal phase does not match expected",
				)
				return
			}
		case err := <-errCh:
			t.Fatalf("error encountered watching job status: %s", err)
		case <-timer.C:
			t.Fatal("timeout waiting for job to reach a terminal phase")
		}
	}
}

func assertJobLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	tc testcase,
) {
	assertLogs(
		t,
		ctx,
		client,
		e,
		&core.LogsSelector{Job: testJobName},
		tc.expectedJobLogsContains,
	)
}

func assertLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	selector *core.LogsSelector,
	contains string,
) {
	if contains == "" {
		return
	}

	logEntryCh, errCh, err :=
		client.Core().Events().Logs().Stream(
			ctx,
			e.ID,
			selector,
			&core.LogStreamOptions{},
		)
	require.NoError(t, err, "error acquiring log stream")

	for {
		select {
		case logEntry, ok := <-logEntryCh:
			if ok {
				if strings.Contains(logEntry.Message, contains) {
					return
				}
			} else {
				logEntryCh = nil
			}
		case err, ok := <-errCh:
			if ok {
				t.Fatalf("error from log stream: %s", err)
			}
			errCh = nil
		case <-ctx.Done():
			break
		}

		// log and err channels empty; we haven't found what we're looking for
		if logEntryCh == nil && errCh == nil {
			t.Fatalf("logs do not contain expected string %q", contains)
		}
	}
}
