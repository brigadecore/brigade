---
title: Gateways
description: How gateways work and how to create your own.
section: operators
weight: 4
aliases:
  - /gateways.md
  - /topics/gateways.md
  - /topics/operators/gateways.md
---

# Brigade Gateways

This guide explains how gateways work, and provides guidance for creating your
own.

## What Is A Brigade Gateway?

The [Brigade architecture](/topics/design) is oriented around the concept that
Brigade scripts run as a response to one or more events. In Brigade, a gateway
is an entity that generates events. Often times, it translates some external
trigger into a Brigade event.

While there is no default gateway that ships with Brigade, installing one
alongside Brigade is as simple as a [Helm] chart install, along with any
additional setup particular to a given gateway. See
[below](#available-gateways) for a current listing of v2-compatible gateways.

All of these provide HTTP(S)-based listeners that receive incoming requests
(from Github or other platforms and systems) and generate Brigade events as a
result.

However, Brigade's gateway system works with more than just webhooks.

For example, the `brig` client behaves similarly to a gateway. When you execute
a `brig event create` command, `brig` creates a Brigade event of source
`brigade.sh/cli` and of type `exec`. Brigade itself processes this event no
differently than it processes events from gateways tracking external activity.

There are no rules about what can be used as a trigger for an event. One could
write a gateway that listens on a message queue, or runs as a chat bot, or
watches files on a filesystem... any of these could be used to trigger a new
Brigade event.

The remainder of this guide explains how gateways work and how you can create
custom gateways.

[Helm]: https://helm.sh

## Available Gateways

Currently, the list of official Brigade v2 gateways that process external
events is as follows:

* [Github Gateway](https://github.com/brigadecore/brigade-github-gateway)
* [BitBucket Gateway](https://github.com/brigadecore/brigade-bitbucket-gateway/tree/v2)
* [CloudEvents Gateway](https://github.com/brigadecore/brigade-cloudevents-gateway)
* [Docker Hub Gateway](https://github.com/brigadecore/brigade-dockerhub-gateway)
* [Azure Container Registry Gateway](https://github.com/brigadecore/brigade-acr-gateway)

Follow the installation instructions provided in each gateway's repository to
learn how to get started.

### The Anatomy of a Brigade Event

All gateways perform the same job, that is, to translate activity and context
from some source into a Brigade event. Let's now look at the structure of a
Brigade event.

A Brigade Event is defined primarily by its source and type values, worker
configuration and worker status.

Here is a YAML representation of an event created via the `brig event create`
command for the [01-hello-world sample project].

```yaml
apiVersion: brigade.sh/v2-beta
kind: Event
metadata:
  created: "2021-08-11T22:22:41.366Z"
  id: 48c960eb-5823-46d0-8390-ec6a2a966b98
projectID: hello-world
source: brigade.sh/cli
type: exec
worker:
  spec:
    configFilesDirectory: .brigade
    defaultConfigFiles:
      brigade.js: |
        console.log("Hello, World!");
    logLevel: DEBUG
    useWorkspace: false
    workspaceSize: 10Gi
  status:
    apiVersion: brigade.sh/v2-beta
    ended: "2021-08-11T22:22:49Z"
    kind: WorkerStatus
    phase: SUCCEEDED
    started: "2021-08-11T22:22:41Z"
```

Let's look at the high-level sections in the event definition above.  They are:

  1. Event metadata, including:
    i. The `apiVersion` of the schema for this event
    ii. The schema `kind`, which will always be `Event`
    iii. The `id` of the Event
    iv. A `created` timestamp for the Event
    v. The `projectID` that the Event is associated with
    vi. The event `source`
    vii. The event `type`
  2. The `worker.spec` section, which contains worker configuration inherited
    from the project definition associated with the event in combination with
    system-level defaults.
  3. The `worker.status` section, which contains the `started` and `ended`
    timestamps and current `phase` of the worker handling the event. In the
    example above, it has already reached the terminal phase of `SUCCEEDED`.

To explore the SDK definitions of an Event object, see the [Go SDK Event] and
[JavaScript/TypeScript SDK Event].

[01-hello-world sample project]: https://github.com/brigadecore/brigade/tree/v2/examples/01-hello-world
[Go SDK Event]: https://github.com/brigadecore/brigade/blob/v2/sdk/v2/core/events.go
[JavaScript/TypeScript SDK Event]: https://github.com/brigadecore/brigade-sdk-for-js/blob/master/src/core/events.ts

## Creating Custom Gateways

Given the above description of how gateways work, we can walk through how a
minimal example gateway can be built. We'll focus on the event creation side
of a Brigade gateway, rather than going over other common attributes, such as
an http(s) server that awaits external webhook events.

Since the Brigade API server is the point of contact for gateway
authentication/authorization and event submission, gateway developers will need
to pick an [SDK] to use. For this example, we'll be using the [Go SDK]. As of
this writing, a [Javascript/Typescript SDK] and a [Rust SDK] (WIP) also exist.

[SDK]: https://github.com/brigadecore/brigade/tree/v2#sdks
[Go SDK]: https://github.com/brigadecore/brigade/tree/v2/sdk
[Javascript/Typescript SDK]: https://github.com/brigadecore/brigade-sdk-for-js
[Rust SDK]: https://github.com/brigadecore/brigade-sdk-for-rust

### Example Gateway

The following example assumes a running Brigade instance deployed with the
default root user enabled. If you'd like to follow along and haven't yet
deployed Brigade, check out the [QuickStart].

[QuickStart]: /intro/quickstart

#### Authentication

As a first course of action for our example gateway written in Go, we'll need
to authenticate with Brigade. Here is what the Go code looks like, with in-line
comments describing each section:

```go
package main

import (
	"context"
	"log"

	brigadeOS "github.com/brigadecore/brigade-foundations/os"
	"github.com/brigadecore/brigade/sdk/v2/authn"
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
}
```

We now have the code needed for the example gateway to communicate with the
Brigade API Server and acquire an authentication token via a root user session.

#### Event Creation

Let's add logic to the gateway so that it can create an event for a provided
Brigade project. We'll focus first on the additions needed for creating an
event in a new function, `createEvent`, that we'll tie together with the `main`
function afterwards.

```go
// createEvent creates a Brigade Event for the provided project, using the
// provided sdk.Client
func createEvent(client sdk.APIClient, projectID string) error {
	ctx := context.Background()

	// Construct a Brigade Event
	event := core.Event{
    // This is the ID/name of the project that the event will be intended for
		ProjectID: projectID,
    // This is the source value for this event
		Source:    "brigade.sh/example-gateway",
    // This is the event's type
		Type:      "hello",
    // This is the event's payload
		Payload:   "Dolly",
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
```

#### Final version

Now that we have our `createEvent` function, we can tie it all together by
adding a bit of logic into `main` to acquire the two values needed by
`createEvent`: the project ID value will be parsed from the command line and
the `sdk.APIClient` object will be constructed using the auth token received
from the API server.

```go
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

	// Get the project ID that the event is destined for from the command line argument
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
		Type: "hello",
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
```

#### Subscribing a project to events from the example gateway

In order to utilize events from the example gateway, we'll need a Brigade
project that subscribes to the corresponding event source
(`brigade.sh/example-gateway`) and event type (`create-event`). We'll also
define an event handler that handles these events and utilizes the attached
payload.

Here's the project definition file.  Note the `spec.eventSubscriptions` section
and the default `brigade.ts` script which contains our event handler:

```yaml
apiVersion: brigade.sh/v2-beta
kind: Project
metadata:
  id: example-gateway-project
description: |-
  An example project that subscribes to events from an example gateway
spec:
  eventSubscriptions:
  - source: brigade.sh/example-gateway
    types:
      - create-event
  workerTemplate:
    logLevel: DEBUG
    defaultConfigFiles:
      brigade.ts: |
        import { events } from "@brigadecore/brigadier"

        events.on("brigade.sh/example-gateway", "create-event", async event => {
          console.log("Hello, " + event.payload + "!")
        })

        events.process()
```

We can save this to `project.yaml` and create it in Brigade via the following
command:

```console
$ brig project create --file project.yaml
```

#### Running the gateway

Now that we have a project subscribing to events from this gateway, we're ready
to build and run the example gateway!

First, we'll need to take care of a few bootstrapping items needed for our Go
program. Here we initialize this program's Go modules file needed for tracking
dependencies. Then, we fetch the needed deps.

```console
$ go mod init
go: creating new go.mod: module github.com/brigadecore/brigade/examples/gateways/example-gateway
go: to add module requirements and sums:
	go mod tidy

$ go mod tidy
go: finding module for package github.com/brigadecore/brigade/sdk/v2/restmachinery
go: finding module for package github.com/brigadecore/brigade/sdk/v2
go: finding module for package github.com/brigadecore/brigade-foundations/os
go: finding module for package github.com/brigadecore/brigade/sdk/v2/authn
go: finding module for package github.com/brigadecore/brigade/sdk/v2/core
go: found github.com/brigadecore/brigade-foundations/os in github.com/brigadecore/brigade-foundations v0.3.0
go: found github.com/brigadecore/brigade/sdk/v2 in github.com/brigadecore/brigade/sdk/v2 v2.0.0-beta.1
go: found github.com/brigadecore/brigade/sdk/v2/authn in github.com/brigadecore/brigade/sdk/v2 v2.0.0-beta.1
go: found github.com/brigadecore/brigade/sdk/v2/core in github.com/brigadecore/brigade/sdk/v2 v2.0.0-beta.1
go: found github.com/brigadecore/brigade/sdk/v2/restmachinery in github.com/brigadecore/brigade/sdk/v2 v2.0.0-beta.1
```

Then, we're ready to run our program, providing the project name as the
argument:

```console
 $ go run main.go example-gateway-project
Event created with ID 46a40cff-0689-466a-9cab-05f4bb9ef9f1 for project example-gateway-project
```

Finally, we can inspect the logs to verify the event was processed by the
worker successfully and that the event payload came through:

```console
$ brig event logs --id 46a40cff-0689-466a-9cab-05f4bb9ef9f1

2021-08-13T22:10:12.726Z INFO: brigade-worker version: 0d7546a
2021-08-13T22:10:12.732Z DEBUG: writing default brigade.ts to /var/vcs/.brigade/brigade.ts
2021-08-13T22:10:12.733Z DEBUG: using npm as the package manager
2021-08-13T22:10:12.733Z DEBUG: path /var/vcs/.brigade/node_modules/@brigadecore does not exist; creating it
2021-08-13T22:10:12.734Z DEBUG: polyfilling @brigadecore/brigadier with /var/brigade-worker/brigadier-polyfill
2021-08-13T22:10:12.734Z DEBUG: compiling brigade.ts with flags --target ES6 --module commonjs --esModuleInterop
2021-08-13T22:10:16.433Z DEBUG: running node brigade.js
Hello, Dolly!
```

#### Wrapping up

Hopefully this brief guide showing a sample gateway written using Brigade's Go
SDK was helpful. All of the sample code can be found in the
[examples/gateways/example-gateway] directory.

We look forward to seeing the Brigade Gateway ecosystem expand with
contributions from readers like you!

[examples/gateways]: https://github.com/brigadecore/brigade/tree/v2/examples/gateways/example-gateway