package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/ghodss/yaml"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/pkg/errors"
)

type testcase struct {
	name          string
	projectFile   string
	terminalPhase core.WorkerPhase
}

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

func main() {
	for _, tc := range testcases {
		err := runTest(tc)
		if err != nil {
			log.Fatalf("Test %q failed: %s", tc.name, err)
		}
	}
}

func runTest(tc testcase) error {
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
	filename := fmt.Sprintf("../testdata/%s", tc.projectFile)

	log.Printf("Running test %q...\n", tc.name)

	projectBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "error reading project file %s", filename)
	}

	if projectBytes, err = yaml.YAMLToJSON(projectBytes); err != nil {
		return errors.Wrapf(err, "error converting file %s to JSON", filename)
	}

	project := core.Project{}
	if err = json.Unmarshal(projectBytes, &project); err != nil {
		return errors.Wrapf(err, "error unmarshaling project file %s", filename)
	}

	authClient := authx.NewSessionsClient(
		apiServerAddress,
		"",
		apiClientOpts,
	)

	token, err := authClient.CreateRootSession(ctx, rootPassword)
	if err != nil {
		return err
	}
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

	if _, err := client.Core().Projects().CreateFromBytes(
		ctx,
		projectBytes,
	); err != nil {
		return err
	}

	event := core.Event{
		ProjectID: project.ID,
		Source:    "github.com/brigadecore/brigade/cli",
		Type:      "exec",
	}

	eList, err := client.Core().Events().Create(ctx, event)
	if err != nil {
		return err
	}

	if len(eList.Items) == 0 {
		return errors.New("event list is empty")
	}
	e := eList.Items[0]

	logEntryCh, errCh, err :=
		client.Core().Events().Logs().Stream(
			ctx,
			e.ID,
			&core.LogsSelector{Container: "vcs"},
			&core.LogStreamOptions{})
	if err != nil {
		return err
	}

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
				return err
			}
			errCh = nil
		case <-ctx.Done():
			return nil
		}

		// log and err channels empty, let's check worker status
		if logEntryCh == nil && errCh == nil {
			for {
				e, err := client.Core().Events().Get(ctx, e.ID)
				if err != nil {
					return err
				}

				phase := e.Worker.Status.Phase
				if phase.IsTerminal() {
					if phase != tc.terminalPhase {
						return fmt.Errorf("worker's terminal phase %q does not match expected %q", phase, tc.terminalPhase)
					}
					return nil
				}
			}
			return nil
		}
	}
	return nil
}
