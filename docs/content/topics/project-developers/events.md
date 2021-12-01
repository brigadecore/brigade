---
title: Events
description: Handling Events in Brigade
section: project-developers
weight: 2
aliases:
  - /events
  - /topics/events.md
  - /topics/project-developers/events.md
---

# Brigade Events

Events are the _lingua franca_ in Brigade: [Gateways] emit events and [Project]
developers write logic to handle them.

In this document, we'll look at:

  * The [structure of an event]
  * How to [configure event subscriptions] on a project
  * How to [handle events] in a Brigade script.

[Gateways]: /topics/operators/gateways
[Project]: /topics/project-developers/projects
[structure of an event]: #event-structure
[configure event subscriptions]: #event-subscriptions
[handle events]: #handling-events

## Event Structure

A Brigade event is defined primarily by its source and type values. These two
must be provided every time an event is emitted into Brigade. However, there
are plenty of other fields that event creators can harness to add additional
context and information for utilization in handlers.

The full list of fields on an Event is:

  * [Source](#source)
  * [Type](#type)
  * [Payload](#payload)
  * [ProjectID](#project-id)
  * [Labels](#labels)
  * [Qualifiers](#qualifiers)
  * [Short Title](#short-title)
  * [Long Title](#long-title)
  * [Source State](#source-state)
  * [Summary](#summary)

Each field is reviewed in depth below.

### Source

An event's Source can be thought of as its domain. For developers writing their
own gateway, the rule of thumb to avoid clashes is to use a URI one controls.
This means leading with one's own domain or the URL for something else one
owns, like the URL for a GitHub repo. As an example, events emitted from
Brigade's [GitHub Gateway] use a source value of `brigade.sh/github`.

### Type

The Type of an event is used to specify increased granularity of the event's
role/purpose as it relates to the source. For example, an event emitted by
the GitHub Gateway with Type `push` signifies that a push event has occurred
on a particular GitHub repo.

### Payload

An event Payload can be used to send input that may be utilized when handling
the event, but is otherwise opaque to Brigade itself. One example would be
sending the original GitHub push event payload on the corresponding GitHub
gateway event, so that these additional details may be parsed in a project's
Brigade script.

### Project ID

Although not normally used by general-purpose gateways, a Project ID value may
be set on an event. In such cases, the event will _only_ be eligible for
receipt by the project indicated by this fields' value. As an example, the
`brig project create --project <id>` command emits an event with this field
set to the supplied project ID value.

### Labels

Projects can choose to utilize Labels for filtering purposes. The labels on a
project's event subscription do not need to exactly match the labels on an
Event in order for the project to receive it. Labels can optionally be used to
narrow an event subscription's scope by selecting only events that are labeled
in a particular way. For example, the Slack gateway labels all events with the
channel of origin. A project's subscription can optionally ignore the label and
get all events for the workspace or specify the label and get events only from
a specific channel.

### Qualifiers

Qualifiers are like labels; albeit _required_ rather than optional. When an
event contains qualifiers, a project's event subscription must declare the same
qualifiers in order for the project to receive it. For example, no one wants to
subscribe to events from _all_ GitHub repositories (or even just all the
repositories in one's org), so the GitHub gateway _qualifies_ events,
specifying the repository of origin. Projects must match this qualifier in
their event subscription in order to receive an event.

### Short Title

_Reserved for future use._ A short title may be provided for an event. This can
then be utilized for logging, categorization or visual representation purposes,
etc.

### Long Title

_Reserved for future use._ A longer, more descriptive title may be provided
for an event. This may be helpful for providing additional context for users
consuming event details.

### Source State

Source State for an event is a key/value map representing event state that can
be persisted by the Brigade API server so that gateways can track event
handling progress and perform other actions, such as updating upstream
services. For example, the GitHub Gateway may utilize this field for tracking
completion of a CI run on a pull request, in order to update GitHub with logs
and pass/fail status.

### Summary

Whereas an event payload represents arbitray information sent inbound to the
client handling the event, the event summary is meant to relay outbound
context/details generated during handling of the original event. The gateway
responsible for emitting the event then has access to a summary of the work
done while processing this event.

To explore the SDK definitions of an Event object, see the [Go SDK Event] and
[JavaScript/TypeScript SDK Event].

[GitHub Gateway]: https://github.com/brigadecore/brigade-github-gateway
[Go SDK Event]: https://github.com/brigadecore/brigade/blob/main/sdk/v2/core/events.go
[JavaScript/TypeScript SDK Event]: https://github.com/brigadecore/brigade-sdk-for-js/blob/main/src/core/events.ts

## Event Subscriptions

In order to receive events, a Brigade project must explicitly subscribe to
them. This is done via an Event Subscription configured on the project.

For an overview of a project definition in Brigade, see the [Projects] doc.

Here we'll look at a sample project and dig into its `eventSubscriptions`
section:

```yaml
apiVersion: brigade.sh/v2
kind: Project
metadata:
  id: hello-world
description: Demonstrates responding to an event with brigadier
spec:
  eventSubscriptions:
  - source: brigade.sh/cli
    types:
    - exec
```

The `hello-world` project above subscribes to one event, that of source
`brigade.sh/cli` and type `exec`. This is the event emitted from the
`brig event create` command.

Additional events and/or additional types of a given event are added under this
`eventSubscriptions` section. [Qualifiers] and [Labels] may also be added, via
the `qualifiers` and `labels` fields respectively, to further refine inbound
events.

Here's a look at an event subscription configuration using all of the above:

```yaml
eventSubscriptions:
- source: brigade.sh/github
  qualifiers:
    repo: example-org/example-repo
  labels:
    branch: main
  types:
  - push
```

[Projects]: /topcs/project-developers/projects
[Qualifiers]: #qualifiers
[Labels]: #labels

## Handling Events

Events that successfully reach a subscribed project can be handled in the
project's Brigade script. Let's revisit the `hello-world` project definition
from above and add an inline `brigade.js` script via the `defaultConfigFiles`
section of its Worker spec. In the script, we'll look at the event handler for
inbound events of source `brigade.sh/cli` and type `exec`:

```yaml
apiVersion: brigade.sh/v2
kind: Project
metadata:
  id: hello-world
description: Demonstrates responding to an event with brigadier
spec:
  eventSubscriptions:
  - source: brigade.sh/cli
    types:
    - exec
  workerTemplate:
    defaultConfigFiles:
      brigade.js: |
        const { events } = require("@brigadecore/brigadier");

        events.on("brigade.sh/cli", "exec", async event => {
          console.log("Hello, World!");
        });

        events.process();
```

The following section of Javascript comprises the event handler:

```javascript
events.on("brigade.sh/cli", "exec", async event => {
  console.log("Hello, World!");
});
```

In fact, all event handlers follow the same pattern, that of:

```javascript
events.on("event source", "event type", async event => {
  // handle event
});
```

Inside the handler, you'll have full access to the `event` object, including
many of the same fields mentioned in the [Event Structure] section above, e.g.
`event.payload`, `event.type`, `event.shortTitle`, etc.

Note that the project script does not need to be embedded in the project
definition as we've done above. It can exist separately, such as in a
git repository. For more scripting techniques and examples, check out the
[Scripting Guide] or peruse the [Example Projects].

[Event Structure]: #event-structure
[Scripting Guide]: /topics/scripting/index.md
[Example Projects]: https://github.com/brigadecore/brigade/tree/main/examples