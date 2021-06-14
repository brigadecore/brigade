package tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/stretchr/testify/require"
)

type testcase struct {
	name              string
	postProjectCreate func(context.Context, sdk.APIClient) error
	project           core.Project
	configFiles       map[string]string
	assertions        func(
		t *testing.T,
		ctx context.Context,
		client sdk.APIClient,
		createdEvents core.EventList,
	)
}

var TestCases = []testcase{
	{
		name: "GitHub - no ref",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-no-ref",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
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
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
	},
	{
		name: "GitHub - full ref",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-full-ref",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
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
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
	},
	{
		name: "GitHub - casual ref",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-casual-ref",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
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
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
	},
	{
		name: "GitHub - commit sha",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-sha",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
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
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
	},
	{
		name: "GitHub - submodules",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-submodules",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
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
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, ".submodules")
		},
	},
	{
		name: "GitHub - private repo",
		postProjectCreate: func(ctx context.Context, client sdk.APIClient) error {
			return client.Core().Projects().Secrets().Set(
				ctx,
				"github-private-ssh",
				core.Secret{
					Key:   "gitSSHKey",
					Value: os.Getenv("BRIGADE_CI_PRIVATE_REPO_SSH_KEY"),
				},
			)
		},
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-private-ssh",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
				WorkerTemplate: core.WorkerSpec{
					Git: &core.GitConfig{
						CloneURL: "git@github.com:brigadecore/private-test-repo.git",
						Ref:      "main",
					},
				},
			},
		},
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseSucceeded)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseSucceeded)
			assertJobLogs(t, ctx, client, e, testJobName, "README.md")
		},
	},
	{
		name: "GitHub - vcs failure",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "github-vcs-fail",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
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
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
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
			ObjectMeta: meta.ObjectMeta{
				ID: "github-job-fail",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
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
				job.primaryContainer.arguments = [
					"-c",
					"'echo Goodbye World && exit 1'"
				]
				await job.run()
			})

			events.process()
		`, testJobName)},
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseFailed)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseFailed)
			assertJobLogs(t, ctx, client, e, testJobName, "Goodbye World")
		},
	},
	{
		name: "Job times out",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "job-times-out",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
			},
		},
		configFiles: map[string]string{
			"brigade.ts": fmt.Sprintf(`
				import { events, Job } from "@brigadecore/brigadier"

				events.on("brigade.sh/cli", "exec", async event => {
					let job = new Job("%s", "alpine", event)
					job.primaryContainer.command = ["sleep"]
					job.primaryContainer.arguments = ["2"]
					job.timeoutSeconds = 1.005
					await job.run()
				})

				events.process()
			`, testJobName)},
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseFailed)
			assertWorkerLogs(t, ctx, client, e, "brigade-worker version")
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseTimedOut)
		},
	},
	{
		name: "Worker times out",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "worker-times-out",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: DefaultEventSubscriptions,
				WorkerTemplate: core.WorkerSpec{
					// Timeout a bit "long" to allow for job to spin up
					TimeoutDuration: "10s",
				},
			},
		},
		configFiles: map[string]string{
			"brigade.ts": fmt.Sprintf(`
				import { events, Job } from "@brigadecore/brigadier"

				events.on("brigade.sh/cli", "exec", async event => {
					let job = new Job("%s", "alpine", event)
					job.primaryContainer.command = ["sleep"]
					job.primaryContainer.arguments = ["60"]
					await job.run()
				})

				events.process()
			`, testJobName)},
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			events core.EventList,
		) {
			require.Len(t, events.Items, 1)
			e := events.Items[0]
			assertWorkerPhase(t, ctx, client, e, core.WorkerPhaseTimedOut)
			assertJobPhase(t, ctx, client, e, testJobName, core.JobPhaseAborted)
		},
	},
	{
		name: "Project not subscribed",
		project: core.Project{
			ObjectMeta: meta.ObjectMeta{
				ID: "project-not-subscribed",
			},
			Spec: core.ProjectSpec{
				EventSubscriptions: []core.EventSubscription{},
			},
		},
		assertions: func(
			t *testing.T,
			ctx context.Context,
			client sdk.APIClient,
			events core.EventList,
		) {
			require.Len(t, events.Items, 0)
		},
	},
}
