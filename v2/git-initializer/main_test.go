// +build integration

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/require"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

type testcase struct {
	name          string
	projectFile   string
	terminalPhase core.WorkerPhase
}

func TestMain(t *testing.T) {
	// TODO: convert to in-line project definitions
	var testcases = []testcase{
		{
			name:          "GitHub - Public Repo - https",
			projectFile:   "github-public-https.yaml",
			terminalPhase: core.WorkerPhaseSucceeded,
		},
		// TODO: once we have mech for injecting private ssh key in secure manner
		// {
		// 	name:          "GitHub - Private Repo - ssh",
		// 	projectFile:   "github-private-ssh.yaml",
		// 	terminalPhase: core.WorkerPhaseSucceeded,
		// },
		{
			name:          "GitHub - Public Repo - submodules",
			projectFile:   "github-init-submodules.yaml",
			terminalPhase: core.WorkerPhaseSucceeded,
		},
		{
			name:          "GitHub - Malformed URL",
			projectFile:   "github-malformed-url.yaml",
			terminalPhase: core.WorkerPhaseFailed,
		},
	}

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
			filename := fmt.Sprintf("./testdata/%s", tc.projectFile)

			projectBytes, err := ioutil.ReadFile(filename)
			require.NoError(t, err, "error reading project file")

			projectBytes, err = yaml.YAMLToJSON(projectBytes)
			require.NoError(t, err, "error converting file to JSON")

			project := core.Project{}
			err = json.Unmarshal(projectBytes, &project)
			require.NoError(t, err, "error unmarshaling project file")

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
			_ = client.Core().Projects().Delete(ctx, project.ID)

			_, err = client.Core().Projects().CreateFromBytes(ctx, projectBytes)
			require.NoError(t, err, "error creating project")

			event := core.Event{
				ProjectID: project.ID,
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
						}
					}
				}
			}
		})
	}
}
