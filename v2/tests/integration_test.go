//go:build integration
// +build integration

package tests

import (
	"context"
	"os"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
	"github.com/brigadecore/brigade/sdk/v2/system"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	ctx := context.Background()

	apiServerAddress := GetRequiredEnvVar(t, "APISERVER_ADDRESS")
	rootPassword := GetRequiredEnvVar(t, "APISERVER_ROOT_PASSWORD")

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

	// Check ping endpoint for expected version
	wantVersion := os.Getenv("VERSION")
	require.NotEmpty(
		t,
		wantVersion,
		"expected the VERSION env var to be non-empty",
	)

	systemClient := system.NewAPIClient(
		apiServerAddress,
		tokenStr,
		apiClientOpts,
	)
	resp, err := systemClient.Ping(ctx)
	require.NoError(t, err)
	require.Equal(
		t,
		wantVersion,
		string(resp.Version),
		"ping response did not match expected",
	)

	// Create the api client for use in tests below
	client := sdk.NewAPIClient(
		apiServerAddress,
		tokenStr,
		apiClientOpts,
	)

	for _, tc := range TestCases {
		// Skip private repo test if required env var not present
		if tc.name == "GitHub - private repo" &&
			os.Getenv("BRIGADE_CI_PRIVATE_REPO_SSH_KEY") == "" {
			continue
		}

		t.Run(tc.name, func(t *testing.T) {
			// Update the project with defaults, if needed
			if len(tc.configFiles) > 0 {
				tc.project.Spec.WorkerTemplate.DefaultConfigFiles = tc.configFiles
			} else {
				tc.project.Spec.WorkerTemplate.DefaultConfigFiles = DefaultConfigFiles
			}

			// Create the test project
			_, err = client.Core().Projects().Create(ctx, tc.project)
			require.NoError(t, err, "error creating project")
			// Run post-project create logic, if defined
			if tc.postProjectCreate != nil {
				require.NoError(t, tc.postProjectCreate(ctx, client))
			}

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

			tc.assertions(t, ctx, client, eList)

			// Delete the test project
			err = client.Core().Projects().Delete(ctx, tc.project.ID)
			require.NoError(t, err, "error deleting project")
		})
	}
}
