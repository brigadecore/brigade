//go:build integration
// +build integration

package tests

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	foundOS "github.com/brigadecore/brigade-foundations/os"
	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/authn"
	"github.com/brigadecore/brigade/sdk/v3/core"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/brigadecore/brigade/sdk/v3/restmachinery"
	"github.com/stretchr/testify/require"
)

var client sdk.APIClient

const testTimeout = 5 * time.Minute

func TestMain(m *testing.M) {
	ctx := context.Background()

	apiServerAddress, err := foundOS.GetRequiredEnvVar("APISERVER_ADDRESS")
	if err != nil {
		log.Fatal(err)
	}
	rootPassword, err := foundOS.GetRequiredEnvVar("APISERVER_ROOT_PASSWORD")
	if err != nil {
		log.Fatal(err)
	}

	apiClientOpts := &restmachinery.APIClientOptions{
		AllowInsecureConnections: true,
	}

	token, err := authn.NewSessionsClient(
		apiServerAddress,
		"",
		apiClientOpts,
	).CreateRootSession(ctx, rootPassword, nil)
	if err != nil {
		log.Fatal(err)
	}

	client = sdk.NewAPIClient(
		apiServerAddress,
		token.Value,
		apiClientOpts,
	)

	os.Exit(m.Run())
}

func TestIntegration(t *testing.T) {
	ctx := context.Background()

	// Check ping endpoint for expected version
	expectedVersion, err := foundOS.GetRequiredEnvVar("VERSION")
	require.NoError(t, err)
	pingResponse, err := client.System().Ping(ctx, nil)
	require.NoError(t, err)
	require.Equal(t, expectedVersion, pingResponse.Version)

	// Create projects used by all test cases
	for _, testCase := range testCases {
		if testCase.shouldTest != nil && !testCase.shouldTest(t) {
			continue
		}
		_, err = client.Core().Projects().Create(ctx, testCase.project, nil)
		require.NoErrorf(t, err, "error creating project %q", testCase.project.ID)
		// nolint: errcheck
		defer client.Core().Projects().Delete(ctx, testCase.project.ID, nil)
		if testCase.postProjectCreate != nil {
			require.NoErrorf(
				t,
				testCase.postProjectCreate(ctx),
				"error running post-create steps for project %q",
				testCase.project.ID,
			)
		}
	}

	// The scheduler learns about new projects every 30 seconds. This grace period
	// ensures the scheduler is listening on every new project's queue before we
	// move on.
	<-time.After(30 * time.Second)

	for _, testCase := range testCases {
		if testCase.shouldTest != nil && !testCase.shouldTest(t) {
			continue
		}
		t.Run(testCase.project.Description, func(t *testing.T) {
			// Verify there are no events for this project
			events, err := client.Core().Events().List(
				ctx,
				&core.EventsSelector{ProjectID: testCase.project.ID},
				&meta.ListOptions{Limit: 1},
			)
			require.NoError(t, err)
			require.Empty(t, events.Items)

			// Create a new event
			event := core.Event{
				ProjectID: testCase.project.ID,
				Source:    "brigade.sh/cli",
				Type:      "exec",
			}

			events, err = client.Core().Events().Create(ctx, event, nil)
			require.NoError(t, err)
			testCase.assertions(t, ctx, events)
		})
	}
}

func assertWorkerPhase(
	t *testing.T,
	ctx context.Context, // nolint: revive
	eventID string,
	expectedPhase core.WorkerPhase,
) {
	statusCh, errCh, err := client.Core().Events().Workers().WatchStatus(
		ctx,
		eventID,
		nil,
	)
	require.NoError(t, err)

	timer := time.NewTimer(testTimeout)
	defer timer.Stop()

	for {
		select {
		case status := <-statusCh:
			phase := status.Phase
			if phase.IsTerminal() {
				require.Equal(
					t,
					expectedPhase,
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
	ctx context.Context, // nolint: revive
	eventID string,
	jobName string,
	expectedPhase core.JobPhase,
) {
	statusCh, errCh, err := client.Core().Events().Workers().Jobs().WatchStatus(
		ctx,
		eventID,
		jobName,
		nil,
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
					expectedPhase,
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

func assertLogs(
	t *testing.T,
	ctx context.Context, // nolint: revive
	eventID string,
	selector *core.LogsSelector,
	expectedLogs string,
) {
	if expectedLogs == "" {
		return
	}

	logEntryCh, errCh, err :=
		client.Core().Events().Logs().Stream(
			ctx,
			eventID,
			selector,
			&core.LogStreamOptions{},
		)
	require.NoError(t, err, "error acquiring log stream")

	for {
		select {
		case logEntry, ok := <-logEntryCh:
			if ok {
				if strings.Contains(logEntry.Message, expectedLogs) {
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
			return
		}

		// log and err channels empty; we haven't found what we're looking for
		if logEntryCh == nil && errCh == nil {
			t.Fatalf("logs do not contain expected string %q", expectedLogs)
		}
	}
}
