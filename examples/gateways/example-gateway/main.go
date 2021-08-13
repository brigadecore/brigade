package main

import (
	"context"
	"fmt"
	"log"
	"os"

	brigadeOS "github.com/brigadecore/brigade-foundations/os"
	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/authn"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

func main() {
	ctx := context.Background()

	// Get the Brigade API server address and root user password from env
	apiServerAddress, err := brigadeOS.GetRequiredEnvVar("APISERVER_ADDRESS")
	if err != nil {
		log.Fatal(err)
	}
	rootPassword, err := brigadeOS.GetRequiredEnvVar("APISERVER_ROOT_PASSWORD")
	if err != nil {
		log.Fatal(err)
	}

	// Get the project ID that the event is destined for from command arg
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalf("Expected one argument, the project ID.")
	}
	projectID := args[0]

	// Create a root session to acquire an auth token

	// The default Brigade deployment mode uses self-signed certs
	// Hence, we allow insecure connections in our APIClientOptions
	apiClientOpts := &restmachinery.APIClientOptions{
		AllowInsecureConnections: true,
	}

	// Create a new auth client
	authClient := authn.NewSessionsClient(
		apiServerAddress,
		"",
		apiClientOpts,
	)

	// Create a root session, which returns an auth token
	token, err := authClient.CreateRootSession(ctx, rootPassword)
	if err != nil {
		log.Fatal(err)
	}
	tokenStr := token.Value

	// Create an API client with the auth token value
	client := sdk.NewAPIClient(
		apiServerAddress,
		tokenStr,
		apiClientOpts,
	)

	// Call the createEvent function with the sdk.APIClient and projectID
	if err := createEvent(client, projectID); err != nil {
		log.Fatal(err)
	}
}

// createEvent creates a Brigade Event for the provided project, using the
// provided sdk.Client
func createEvent(client sdk.APIClient, projectID string) error {
	ctx := context.Background()

	// Construct a Brigade Event
	event := core.Event{
		// This is the ID/name of the project that the event will be intended for
		ProjectID: projectID,
		// This is the source value for this event
		Source: "brigade.sh/example-gateway",
		// This is the event's type
		Type: "create-event",
		// This is the event's payload
		Payload: "Dolly",
	}

	// Create the Brigade Event
	events, err := client.Core().Events().Create(ctx, event)
	if err != nil {
		return err
	}

	// If the returned events list has no items, no event was created
	if len(events.Items) != 1 {
		fmt.Printf(
			"No event was created. "+
				"Does project %s subscribe to events of source %s and type %s?\n",
			projectID,
			event.Source,
			event.Type,
		)
		return nil
	}

	// The Brigade event was successfully created!
	fmt.Printf(
		"Event created with ID %s for project %s\n",
		events.Items[0].ID,
		projectID,
	)
	return nil
}
