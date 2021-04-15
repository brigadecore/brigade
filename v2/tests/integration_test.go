// +build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/system"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	testJobName = "test-job"
	testTimeout = time.Duration(120 * time.Second)
)

type testcase struct {
	name        string
	project     core.Project
	configFiles map[string]string
	assertions  func(
		t *testing.T,
		ctx context.Context,
		client sdk.APIClient,
		event core.Event,
	)
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
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			e core.Event,
		) {
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
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
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			e core.Event,
		) {
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
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
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			e core.Event,
		) {
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
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
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			e core.Event,
		) {
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
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
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			e core.Event,
		) {
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, ".submodules")
		},
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
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			e core.Event,
		) {
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseFailed)
			assertVCSLogs(
				t,
				ctx,
				client,
				e,
				`reference "non-existent" not found in repo `+
					`"https://github.com/brigadecore/empty-testbed.git"`,
			)
		},
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

			events.on("brigade.sh/cli", "exec", async event => {
				let job = new Job("%s", "alpine", event)
				job.primaryContainer.command = ["sh"]
				job.primaryContainer.arguments = ["-c", "'echo Goodbye World && exit 1'"]
				await job.run()
			})

			events.process()
		`, testJobName)},
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			e core.Event,
		) {
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseFailed)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseFailed)
			assertJobLogs(t, ctx, client, e, testJobName, "Goodbye World")
		},
	},
}

var defaultConfigFiles = map[string]string{
	"brigade.ts": fmt.Sprintf(`
		import { events, Job } from "@brigadecore/brigadier"

		events.on("brigade.sh/cli", "exec", async event => {
			let job = new Job("%s", "alpine", event)
			job.primaryContainer.sourceMountPath = "/var/vcs"
			job.primaryContainer.command = ["ls"]
			job.primaryContainer.arguments = ["-haltr", "/var/vcs"]
			await job.run()
		})

		events.process()
	`, testJobName)}

func TestMain(t *testing.T) {
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

	// Check unversionedPing endpoint for expected version
	wantResp := os.Getenv("VERSION")
	require.NotEmpty(t, wantResp, "expected the VERSION env var to be non-empty")

	systemClient := system.NewAPIClient(
		apiServerAddress,
		tokenStr,
		apiClientOpts,
	)
	resp, err := systemClient.UnversionedPing(ctx)
	require.NoError(t, err)
	require.Equal(t, wantResp, string(resp), "ping response did not match expected")

	// Create the api client for use in tests below
	client := sdk.NewAPIClient(
		apiServerAddress,
		tokenStr,
		apiClientOpts,
	)

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.project.ID = "test-project"

			// Delete the test project (we're sharing the name between tests)
			err = client.Core().Projects().Delete(ctx, tc.project.ID)
			if _, ok := errors.Cause(err).(*meta.ErrNotFound); !ok {
				require.NoError(t, err, "error deleting project")
			}

			// Update the project with defaults, if needed
			if len(tc.configFiles) > 0 {
				tc.project.Spec.WorkerTemplate.DefaultConfigFiles = tc.configFiles
			} else {
				tc.project.Spec.WorkerTemplate.DefaultConfigFiles = defaultConfigFiles
			}

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
				Source:    "brigade.sh/cli",
				Type:      "exec",
			}

			eList, err = client.Core().Events().Create(ctx, event)
			require.NoError(t, err, "error creating event")
			require.Equal(t, 1, len(eList.Items), "event list items should be exactly one")

			tc.assertions(t, ctx, client, eList.Items[0])
		})
	}
}

func assertWorkerPhase(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	wantPhase core.WorkerPhase,
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
					wantPhase,
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

func assertJobPhase(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	job string,
	wantPhase core.JobPhase,
) {
	statusCh, errCh, err := client.Core().Events().Workers().Jobs().WatchStatus(
		ctx,
		e.ID,
		job,
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
					wantPhase,
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

func assertVCSLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	wantLogs string,
) {
	assertLogs(
		t,
		ctx,
		client,
		e,
		&core.LogsSelector{Container: "vcs"},
		wantLogs,
	)
}

func assertWorkerLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	wantLogs string,
) {
	assertLogs(
		t,
		ctx,
		client,
		e,
		&core.LogsSelector{},
		wantLogs,
	)
}

func assertJobLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	job string,
	wantLogs string,
) {
	assertLogs(
		t,
		ctx,
		client,
		e,
		&core.LogsSelector{Job: job},
		wantLogs,
	)
}

func assertLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	selector *core.LogsSelector,
	wantLogs string,
) {
	if wantLogs == "" {
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
				if strings.Contains(logEntry.Message, wantLogs) {
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
			t.Fatalf("logs do not contain expected string %q", wantLogs)
		}
	}
}
