//go:build integration
// +build integration

package tests

import (
	"context"
	"os"
	"testing"

	"github.com/brigadecore/brigade/sdk/v3"
	"github.com/brigadecore/brigade/sdk/v3/meta"
	"github.com/stretchr/testify/require"
)

const testJobName = "test-job"

var testEventSubscriptions = []sdk.EventSubscription{
	{
		Source: "brigade.sh/cli",
		Types: []string{
			"exec",
		},
	},
}

var testConfigFiles = map[string]string{
	"brigade.ts": `
import { events, Job } from "@brigadecore/brigadier"
events.on("brigade.sh/cli", "exec", async event => {
	let job = new Job("test-job", "alpine", event)
	job.primaryContainer.sourceMountPath = "/var/vcs"
	job.primaryContainer.command = ["ls"]
	job.primaryContainer.arguments = ["-haltr", "/var/vcs"]
	await job.run()
})
events.process()`,
}

var testCases = []struct {
	shouldTest        func(*testing.T) bool
	project           sdk.Project
	postProjectCreate func(context.Context) error
	assertions        func(*testing.T, context.Context, sdk.EventList)
}{
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-no-ref",
			},
			Description: "GitHub - no ref",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
					},
					DefaultConfigFiles: testConfigFiles,
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseSucceeded)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseSucceeded)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Job: testJobName},
				"README.md",
			)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-full-ref",
			},
			Description: "GitHub - full ref",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
						Ref:      "refs/heads/main",
					},
					DefaultConfigFiles: testConfigFiles,
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseSucceeded)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseSucceeded)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Job: testJobName},
				"README.md",
			)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-casual-ref",
			},
			Description: "GitHub - casual ref",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
						Ref:      "main",
					},
					DefaultConfigFiles: testConfigFiles,
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseSucceeded)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseSucceeded)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Job: testJobName},
				"README.md",
			)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-sha",
			},
			Description: "GitHub - commit sha",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
						Commit:   "589e15029e1e44dee48de4800daf1f78e64287c0",
					},
					DefaultConfigFiles: testConfigFiles,
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseSucceeded)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseSucceeded)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Job: testJobName},
				"README.md",
			)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-submodules",
			},
			Description: "GitHub - submodules",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL:       "https://github.com/brigadecore/empty-testbed.git",
						InitSubmodules: true,
					},
					DefaultConfigFiles: testConfigFiles,
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseSucceeded)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseSucceeded)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Job: testJobName},
				"README.md",
			)
		},
	},
	{
		shouldTest: func(t *testing.T) bool {
			return os.Getenv("BRIGADE_CI_PRIVATE_REPO_SSH_KEY") != ""
		},
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-private-ssh",
			},
			Description: "GitHub - private repo over ssh",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL: "git@github.com:brigadecore/private-test-repo.git",
						Ref:      "main",
					},
					DefaultConfigFiles: testConfigFiles,
				},
			},
		},
		postProjectCreate: func(ctx context.Context) error {
			return client.Core().Projects().Secrets().Set(
				ctx,
				"github-private-ssh",
				sdk.Secret{
					Key:   "gitSSHKey",
					Value: os.Getenv("BRIGADE_CI_PRIVATE_REPO_SSH_KEY"),
				},
				nil,
			)
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseSucceeded)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseSucceeded)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Job: testJobName},
				"README.md",
			)
		},
	},
	{
		shouldTest: func(t *testing.T) bool {
			return os.Getenv("BRIGADE_CI_PRIVATE_REPO_PAT") != ""
		},
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-private-https",
			},
			Description: "GitHub - private repo over https",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL: "https://github.com/brigadecore/private-test-repo.git",
						Ref:      "main",
					},
					DefaultConfigFiles: testConfigFiles,
				},
			},
		},
		postProjectCreate: func(ctx context.Context) error {
			return client.Core().Projects().Secrets().Set(
				ctx,
				"github-private-https",
				sdk.Secret{
					Key:   "gitPassword",
					Value: os.Getenv("BRIGADE_CI_PRIVATE_REPO_PAT"),
				},
				nil,
			)
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseSucceeded)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseSucceeded)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Job: testJobName},
				"README.md",
			)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-vcs-fail",
			},
			Description: "GitHub - vcs failure",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
						Ref:      "non-existent",
					},
					DefaultConfigFiles: testConfigFiles,
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseFailed)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Container: "vcs"},
				`error resolving ref "non-existent"`,
			)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-job-fail",
			},
			Description: "GitHub - job fails",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					Git: &sdk.GitConfig{
						CloneURL: "https://github.com/brigadecore/empty-testbed.git",
					},
					DefaultConfigFiles: map[string]string{
						"brigade.ts": `
import { events, Job } from "@brigadecore/brigadier"
events.on("brigade.sh/cli", "exec", async event => {
	let job = new Job("test-job", "alpine", event)
	job.primaryContainer.command = ["sh"]
	job.primaryContainer.arguments = [
		"-c",
		"'echo Goodbye World && exit 1'"
	]
	await job.run()
})
events.process()`,
					},
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseFailed)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseFailed)
			assertLogs(
				t,
				ctx,
				event.ID,
				&sdk.LogsSelector{Job: testJobName},
				"Goodbye World",
			)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "job-times-out",
			},
			Description: "Job times out",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					DefaultConfigFiles: map[string]string{
						"brigade.ts": `
import { events, Job } from "@brigadecore/brigadier"
events.on("brigade.sh/cli", "exec", async event => {
	let job = new Job("test-job", "alpine", event)
	job.primaryContainer.command = ["sleep"]
	job.primaryContainer.arguments = ["2"]
	job.timeoutSeconds = 1.005
	await job.run()
})
events.process()`,
					},
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseFailed)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseTimedOut)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "worker-times-out",
			},
			Description: "Worker times out",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					// Timeout a bit "long" to allow for job to spin up
					TimeoutDuration: "10s",
					DefaultConfigFiles: map[string]string{
						"brigade.ts": `
import { events, Job } from "@brigadecore/brigadier"
events.on("brigade.sh/cli", "exec", async event => {
	let job = new Job("test-job", "alpine", event)
	job.primaryContainer.command = ["sleep"]
	job.primaryContainer.arguments = ["60"]
	await job.run()
})
events.process()`,
					},
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseTimedOut)
			assertJobPhase(t, ctx, event.ID, testJobName, sdk.JobPhaseAborted)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "project-not-subscribed",
			},
			Description: "Project not subscribed",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: []sdk.EventSubscription{},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Empty(t, events.Items)
		},
	},
	{
		project: sdk.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "fallible-job-fail",
			},
			Description: "fallible job fails",
			Spec: sdk.ProjectSpec{
				EventSubscriptions: testEventSubscriptions,
				WorkerTemplate: sdk.WorkerSpec{
					DefaultConfigFiles: map[string]string{
						"brigade.ts": `
import { events, Job } from "@brigadecore/brigadier"
events.on("brigade.sh/cli", "exec", async event => {
	let job1 = new Job("test-job1", "alpine", event)
	job1.primaryContainer.command = ["false"]
	job1.fallible = true
	await job1.run()

	let job2 = new Job("test-job2", "alpine", event)
	job2.primaryContainer.command = ["true"]
	await job2.run()
})
events.process()`,
					},
				},
			},
		},
		assertions: func(t *testing.T, ctx context.Context, events sdk.EventList) {
			require.Len(t, events.Items, 1)
			event := events.Items[0]
			assertWorkerPhase(t, ctx, event.ID, sdk.WorkerPhaseSucceeded)
			assertLogs(t, ctx, event.ID, nil, "brigade-worker version")
			assertJobPhase(t, ctx, event.ID, "test-job1", sdk.JobPhaseFailed)
			assertJobPhase(t, ctx, event.ID, "test-job2", sdk.JobPhaseSucceeded)
		},
	},
}
