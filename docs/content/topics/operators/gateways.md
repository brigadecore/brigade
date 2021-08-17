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

## Example Gateway

The following example assumes a running Brigade instance has been deployed and
the ability to create a service account is in place (e.g. you have the role of
'ADMIN' or you are logged in as the root user). If you'd like to follow along
and haven't yet deployed Brigade, check out the [QuickStart].

[QuickStart]: /intro/quickstart

### Preparation

#### Service Account creation

All Brigade gateways require a service account token for authenticating with
Brigade when submitting an event into the system. As preparation then, we'll
create a service account for this gateway and save the generated token for use
in our program.

```console
$ brig service-account create \
	--id example-gateway \
	--description example-gateway
```

Make note of the token returned. This value will be used in another step. It is
your only opportunity to access this value, as Brigade does not save it.

Authorize this service account to create new events:

```console
$ brig role grant EVENT_CREATOR \
    --service-account example-gateway \
    --source brigade.sh/example-gateway
```

Note: The `--source brigade.sh/example-gateway` option specifies that this
service account can be used only to create events having a value of
`brigade.sh/example-gateway` in the event's `source` field. This is a security
measure that prevents the gateway from using this token for impersonating other
gateways.

#### Go setup

We'll be using the [Go SDK] for our example gateway program and we'll need to
do a bit of prep. We're assuming your system has Go installed and configured
properly. (If not, please visit the [Go installation docs] to do so.)

Let's create a directory where our program's `main.go` file can reside:

```console
$ mkdir example-gateway

$ cd example-gateway

$ touch main.go
```

[Go installation docs]: https://golang.org/doc/install

### Example Gateway code

Now we're ready to code! Open `main.go` in the editor of your choice and add in
the following Go code.

The program consists of a `main` function which procures the Brigade API server
address and the gateway token (generated above) via environment variables. It
then constructs an API client from these values and passes this to the
`createEvent` helper function. This function builds a Brigade Event with the
pertinent fields populated and then calls the SDK's event create function.

See the in-line comments for further description around each section.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/brigadecore/brigade/sdk/v2"
	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/brigadecore/brigade/sdk/v2/restmachinery"
)

func main() {
	// Get the Brigade API server address and gateway token from the environment
	apiServerAddress := os.Getenv("APISERVER_ADDRESS")
	if apiServerAddress == "" {
		log.Fatalf("Required environment variable APISERVER_ADDRESS not found.\n")
	}
	gatewayToken := os.Getenv("GATEWAY_TOKEN")
	if gatewayToken == "" {
		log.Fatalf("Required environment variable GATEWAY_TOKEN not found.\n")
	}

	// The default Brigade deployment mode uses self-signed certs
	// Hence, we allow insecure connections in our APIClientOptions
	// This can be changed to false if insecure connections should not be allowed
	apiClientOpts := &restmachinery.APIClientOptions{
		AllowInsecureConnections: true,
	}

	// Create an API client with the gateway token value
	client := sdk.NewAPIClient(
		apiServerAddress,
		gatewayToken,
		apiClientOpts,
	)

	// Call the createEvent function with the sdk.APIClient provided
	if err := createEvent(client); err != nil {
		log.Fatal(err)
	}
}

// createEvent creates a Brigade Event using the provided sdk.Client
func createEvent(client sdk.APIClient) error {
	ctx := context.Background()

	// Construct a Brigade Event
	event := core.Event{
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
		fmt.Println("No event was created.")
		return nil
	}

	// The Brigade event was successfully created!
	fmt.Printf("Event created with ID %s\n", events.Items[0].ID)
	return nil
}
```

Let's briefly look at the Brigade Event object from above.

```go
  // Construct a Brigade Event
  event := core.Event{
    // This is the source value for this event
    Source:    "brigade.sh/example-gateway",
    // This is the event's type
    Type:      "create-event",
    // This is the event's payload
    Payload:   "Dolly",
  }
```

We've filled in the core fields needed for any Brigade event, `Source` and
`Type`. As a bonus, we're also adding a `Payload`. However, that's just the
start of what a Brigade Event can contain. Other notable fields worth
researching are:

- `ProjectID`: When supplied, the event will _only_ be eligible for receipt by
	a specific project.

- `Qualifiers`: A list of qualifier values. For a project to receive an event,
	the qualifiers on an event's subscription must exactly match the qualifiers
	on the event (in addition to matching source and type).

- `Labels`: A list of labels. Projects can choose to utilize these for
  filtering purposes. In contrast to qualifiers, a project's event
	subscription does not need to match an event's labels in order to receive it.
	Labels, however, can be used to narrow an event subscription by optionally
	selecting only events that are labeled in a particular way.

- `ShortTitle`: A short title for the event.

- `LongTitle`: A longer, more descriptive title for the event.

- `SourceState`: A key/value map representing event state that can be persisted
  by the Brigade API server so that gateways can track event handling progress
  and perform other actions, such as updating upstream services.

### Subscribing a project to events from the example gateway

In order to utilize events from the example gateway, we'll need a Brigade
project that subscribes to the corresponding event source
(`brigade.sh/example-gateway`) and event type (`hello`). We'll also
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
      - hello
  workerTemplate:
    logLevel: DEBUG
    defaultConfigFiles:
      brigade.ts: |
        import { events } from "@brigadecore/brigadier"

        events.on("brigade.sh/example-gateway", "hello", async event => {
          console.log("Hello, " + event.payload + "!")
        })

        events.process()
```

We can save this to `project.yaml` and create it in Brigade via the following
command:

```console
$ brig project create --file project.yaml
```

### Running the gateway

Now that we have a project subscribing to events from this gateway, we're ready
to build and run the example gateway!

First, we'll need to take care of a few bootstrapping items needed for our Go
program. Here we initialize this program's Go modules file needed for tracking
dependencies. Then, we fetch the needed Brigade SDK dependency.

```console
$ go mod init example-gateway
go: creating new go.mod: module example-gateway
go: to add module requirements and sums:
	go mod tidy

$ go get github.com/brigadecore/brigade/sdk/v2
go get: added github.com/brigadecore/brigade/sdk/v2 v2.0.0-beta.1
```

Now we're ready to run our program. We export the values required by the
gateway and then run the gateway:

```console
$ export APISERVER_ADDRESS=<Brigade API server address>

$ export GATEWAY_TOKEN=<Brigade service account token from above>

$ go run main.go
Event created with ID 46a40cff-0689-466a-9cab-05f4bb9ef9f1
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

### Wrapping up

Hopefully this brief guide showing a sample gateway written using Brigade's Go
SDK was helpful. All of the sample code can be found in the
[examples/gateways/example-gateway] directory.

We look forward to seeing the Brigade Gateway ecosystem expand with
contributions from readers like you!

[examples/gateways]: https://github.com/brigadecore/brigade/tree/v2/examples/gateways/example-gateway