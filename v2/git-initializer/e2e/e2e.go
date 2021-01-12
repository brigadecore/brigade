package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/ghodss/yaml"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authx"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/pkg/errors"
)

func main() {
	// TODO: iterate over various project files and run tests
	// - ssh
	// - git submodules
	// - other?
	err := runTest()
	if err != nil {
		log.Fatalf("test failed: %s", err)
	}
}

func runTest() error {
	filename := "../testdata/github-demo.yaml"

	projectBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "error reading project file %s", filename)
	}

	if strings.HasSuffix(filename, ".yaml") ||
		strings.HasSuffix(filename, ".yml") {
		if projectBytes, err = yaml.YAMLToJSON(projectBytes); err != nil {
			return errors.Wrapf(err, "error converting file %s to JSON", filename)
		}
	}

	project := core.Project{}
	if err = json.Unmarshal(projectBytes, &project); err != nil {
		return errors.Wrapf(err, "error unmarshaling project file %s", filename)
	}

	ctx := context.Background()
	apiServerAddress := "https://localhost:7000"
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
	// TODO: create/use error types to ignore error if expected
	// e.g. project doesn't yet exist, but still surface unexpected errors
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
			&core.LogsSelector{
				Container: "vcs",
			},
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

		if logEntryCh == nil && errCh == nil {
			// TODO: add timeout
			for {
				e, err := client.Core().Events().Get(ctx, e.ID)
				if err != nil {
					return err
				}

				phase := e.Worker.Status.Phase
				if core.IsTerminal(phase) {
					if phase != core.WorkerPhaseSucceeded {
						return fmt.Errorf("worker's terminal phase: %s", phase)
					}
					return nil
				}
			}
			return nil
		}
	}
	return nil
}
