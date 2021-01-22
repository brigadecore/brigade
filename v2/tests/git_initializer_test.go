// +build integration

package tests

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/stretchr/testify/require"
)

var defaultConfigFiles = map[string]string{
	"brigade.js": `
		const { events } = require("brigadier");

		events.on("github.com/brigadecore/brigade/cli", "exec", () => {
			console.log("Hello, World!");
		});
	`}

var testcases = []struct {
	name          string
	project       core.Project
	terminalPhase core.WorkerPhase
}{
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
		terminalPhase: core.WorkerPhaseSucceeded,
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
		terminalPhase: core.WorkerPhaseSucceeded,
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
		terminalPhase: core.WorkerPhaseSucceeded,
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
		terminalPhase: core.WorkerPhaseSucceeded,
	},
	{
		name: "GitHub - non-existent ref",
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
		terminalPhase: core.WorkerPhaseFailed,
	},
	{
		name: "GitHub - submodules",
		project: core.Project{
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						// TODO: host similar repo in brigadecore
						CloneURL:       "https://github.com/sgoings/makeup.git",
						InitSubmodules: true,
					},
				},
			},
		},
		terminalPhase: core.WorkerPhaseSucceeded,
	},
}

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

			authClient := authx.NewSessionsClient(
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
			tc.project.Spec.WorkerTemplate.DefaultConfigFiles = defaultConfigFiles

			// Delete the project and ignore any errors
			// TODO: create/use error types in core, so we can ignore error if expected
			// e.g. project doesn't yet exist, and surface unexpected errors
			_ = client.Core().Projects().Delete(ctx, tc.project.ID)

			_, err = client.Core().Projects().Create(ctx, tc.project)
			require.NoError(t, err, "error creating project")

			event := core.Event{
				ProjectID: tc.project.ID,
				Source:    "github.com/brigadecore/brigade/cli",
				Type:      "exec",
			}

			eList, err := client.Core().Events().Create(ctx, event)
			require.NoError(t, err, "error creating event")
			require.NotEqual(t, 0, len(eList.Items), "event list is empty")

			e := eList.Items[0]

			logEntryCh, errCh, err :=
				client.Core().Events().Logs().Stream(
					ctx,
					e.ID,
					&core.LogsSelector{Container: "vcs"},
					&core.LogStreamOptions{})
			require.NoError(t, err, "error acquiring log stream")

			for {
				select {
				case logEntry, ok := <-logEntryCh:
					if ok {
						log.Println(logEntry.Message)
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

				// log and err channels empty, let's check worker status
				if logEntryCh == nil && errCh == nil {
					for {
						event, err := client.Core().Events().Get(ctx, e.ID)
						require.NoError(t, err, "error getting event")

						phase := event.Worker.Status.Phase
						if phase.IsTerminal() {
							require.Equal(t, tc.terminalPhase, phase, "worker's terminal phase does not match expected")
							break
						}
					}
					break
				}
			}
		})
	}
}
