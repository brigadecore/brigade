package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/stretchr/testify/require"
)

const (
	testJobName = "test-job"
	testTimeout = time.Duration(120 * time.Second)
)

var (
	DefaultEventSubscriptions = []core.EventSubscription{
		{
			Source: "brigade.sh/cli",
			Types: []string{
				"exec",
			},
		},
	}

	DefaultConfigFiles = map[string]string{
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
)

func GetRequiredEnvVar(t *testing.T, name string) string {
	val := os.Getenv(name)
	if val == "" {
		t.Fatalf(
			"value not found for required environment variable %s",
			name,
		)
	}
	return val
}

// nolint: golint
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

// nolint: golint
func assertJobPhase(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	job string, // nolint: unparam
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

// nolint: golint
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

// nolint: golint
func assertWorkerLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	wantLogs string, // nolint: unparam
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

// nolint: golint
func assertJobLogs(
	t *testing.T,
	ctx context.Context,
	client sdk.APIClient,
	e core.Event,
	job string, // nolint: unparam
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

// nolint: golint
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
