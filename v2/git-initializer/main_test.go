// +build integration

package main

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

var defaultBrigadeJS = `
	const { events } = require("brigadier");

	events.on("github.com/brigadecore/brigade/cli", "exec", () => {
		console.log("Hello, World!");
	});
`

var testcases = []struct {
	name          string
	project       core.Project
	terminalPhase core.WorkerPhase
}{
	{
		name: "GitHub - Public Repo - https",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "public-https",
			},
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "https://github.com/krancour/brigade2-pipeline-demo.git",
					},
					DefaultConfigFiles: map[string]string{
						"brigade.js": defaultBrigadeJS,
					},
				},
			},
		},
		terminalPhase: core.WorkerPhaseSucceeded,
	},
	// TODO: once we have mech for injecting private ssh key in secure manner
	// {
	// 	name:          "GitHub - Private Repo - ssh",
	// 	project: core.Project{
	// 		ObjectMeta: meta.ObjectMeta{
	// 			ID: "private-ssh",
	// 		},
	// 		Spec: core.ProjectSpec{
	// 			WorkerTemplate: core.WorkerSpec{
	// 				Git: &core.GitConfig{
	//					// TODO: host similar repo in brigadecore
	// 					CloneURL: "git@github.com:vdice/testrepo.git",
	// 					Commit: "a292c2ffe0226f702519ec2833054953e82a816e",
	// 				},
	// 				DefaultConfigFiles: map[string]string{
	// 					"brigade.js": defaultBrigadeJS,
	// 				},
	// 			},
	// 		},
	// 	},
	// 	terminalPhase: core.WorkerPhaseSucceeded,
	// },
	{
		name: "GitHub - Public Repo - submodules",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "init-submodules",
			},
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						// TODO: host similar repo in brigadecore
						CloneURL:       "https://github.com/sgoings/makeup.git",
						InitSubmodules: true,
					},
					DefaultConfigFiles: map[string]string{
						"brigade.js": defaultBrigadeJS,
					},
				},
			},
		},
		terminalPhase: core.WorkerPhaseSucceeded,
	},
	{
		name: "GitHub - Malformed URL",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "malformed-url",
			},
			Spec: core.ProjectSpec{
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "https://github.com/brigadecore/brigadee.git",
					},
					DefaultConfigFiles: map[string]string{
						"brigade.js": defaultBrigadeJS,
					},
				},
			},
		},
		terminalPhase: core.WorkerPhaseFailed,
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
