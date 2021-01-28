// +build integration

package tests

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/stretchr/testify/require"
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
			"brigade.ts": `
			import { events, Job } from "@brigadecore/brigadier"

			events.on("github.com/brigadecore/brigade/cli", "exec", async event => {
				let job = new Job("ls", "alpine", event)
				job.primaryContainer.command = ["sh"]
				job.primaryContainer.arguments = ["-c", "'echo Goodbye World && exit 1'"]
				await job.run()
			});
		`},
		expectedJobLogsContains: "Goodbye World",
		expectedJobPhase:        core.JobPhaseFailed,
		expectedWorkerPhase:     core.WorkerPhaseFailed,
	},
}

var defaultConfigFiles = map[string]string{
	"brigade.ts": `
		import { events, Job } from "@brigadecore/brigadier"

		events.on("github.com/brigadecore/brigade/cli", "exec", async event => {
			let job = new Job("ls", "alpine", event)
			job.primaryContainer.sourceMountPath = "/var/vcs"
			job.primaryContainer.command = ["ls"]
			job.primaryContainer.arguments = ["-haltr", "/var/vcs"]
			await job.run()
		});
	`}

func TestMain(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel :=
				context.WithTimeout(context.Background(), 120*time.Second)
			defer cancel()

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
	tc testcase) {

	for {
		event, err := client.Core().Events().Get(ctx, e.ID)
		require.NoError(t, err, "error getting event")

		phase := event.Worker.Status.Phase
		if phase.IsTerminal() {
			require.Equal(t, tc.expectedWorkerPhase, phase, "worker's terminal phase does not match expected")
			break
		}
	}
}

func assertWorkerLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	tc testcase) {

	// Check vcs container logs
	selector := &core.LogsSelector{Container: "vcs"}
	logs := captureLogs(t, ctx, client, e, selector)

	if tc.expectedVCSLogsContains != "" {
		require.Contains(t, logs, tc.expectedVCSLogsContains, "vcs init container logs do not match expected")
	}

	// We expect the vcs init container to fail for this test,
	// therefore, assertions on worker logs are not applicable
	if tc.name == "GitHub - vcs failure" {
		return
	}

	// Check worker container logs
	selector = &core.LogsSelector{}
	logs = captureLogs(t, ctx, client, e, selector)

	if tc.expectedWorkerLogsContains != "" {
		require.Contains(t, logs, tc.expectedWorkerLogsContains, "worker's logs do not match expected")
	} else {
		// Look for default logs that we expect must exist
		require.Contains(t, logs, "brigade-worker version", "worker's logs do not match expected")
	}
}

func assertJobTerminalPhase(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	tc testcase) {

	for {
		event, err := client.Core().Events().Get(ctx, e.ID)
		require.NoError(t, err, "error getting event")

		jobs := event.Worker.Jobs
		require.Equal(t, 1, len(jobs), "expected job count is not 1")

		job := jobs["ls"]
		require.NotNil(t, job, "expected job does not exist")

		phase := job.Status.Phase
		if phase.IsTerminal() {
			require.Equal(t, tc.expectedJobPhase, phase, "job's terminal phase does not match expected")
			break
		}
	}
}

func assertJobLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	tc testcase) {

	selector := &core.LogsSelector{Job: "ls"}
	logs := captureLogs(t, ctx, client, e, selector)

	if tc.expectedJobLogsContains != "" {
		require.Contains(t, logs, tc.expectedJobLogsContains, "job's logs do not match expected")
	}
}

func captureLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	selector *core.LogsSelector) string {

	logEntryCh, errCh, err :=
		client.Core().Events().Logs().Stream(
			ctx,
			e.ID,
			selector,
			&core.LogStreamOptions{},
		)
	require.NoError(t, err, "error acquiring log stream")

	var b bytes.Buffer
	for {
		select {
		case logEntry, ok := <-logEntryCh:
			if ok {
				b.Write([]byte(logEntry.Message))
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

		// log and err channels empty, let's return logs
		if logEntryCh == nil && errCh == nil {
			return b.String()
		}
	}
}
