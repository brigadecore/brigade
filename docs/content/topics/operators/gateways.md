---
title: Gateways
description: How gateways work and how to create your own.
aliases:
  - /gateways.md
  - /topics/gateways.md
  - /topics/operators/gateways.md
---

TODO: update per v2

# Brigade Gateways

This guide explains how gateways work, and provides guidance for creating your own
gateway.

## What Is A Brigade Gateway?

The [Brigade architecture](design.md) is oriented around the concept that Brigade
scripts run as a response to one or more events. In Brigade, a _gateway_ is an
entity that generates events. Often times, it translates some external trigger
into a Brigade event.

Brigade ships with the ability to enable various gateways that are ready to go.

These include the [Container Registry Gateway](dockerhub.md), the [Github Gateway](./github.md)
and the [Generic Gateway](./genericgateway.md).  They can all be enabled via top-level
Helm [chart flags](https://github.com/brigadecore/charts/blob/master/charts/brigade/values.yaml).

All of these provide HTTP-based listeners that receive incoming requests
(from a container registry, Github or other platforms and systems) and generate
Brigade events as a result.

However, Brigade's gateway system works with more than just webhooks.

For example, the `brig` client also acts as a gateway. When you execute a `brig run`
command, `brig` creates a Brigade event. By default, it emits an `exec` event. And
Brigade itself processes this event no differently than it processes the GitHub
or container registry events.

There are no rules about what can be used as a trigger for an event. One could
write a gateway that listens on a message queue, or runs as a chat bot, or watches
files on a filesystem... any of these could be used to trigger a new Brigade event.

The remainder of this guide explains how gateways work and how you can create custom
gateways.

## An Event Is A Secret

The most important thing to understand about a Brigade event is that it is simply
a [Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/)
with special labels and data.

When a new and appropriately labeled secret is created in Kubernetes, the Brigade
controller will read that secret and start a new Brigade worker to handle the event.
Secrets have several characteristics that make them a great fit for this role:

- They are designed to protect data (and we expect them to mature in this capacity)
- They can be mounted as volumes and environment variables.
- The payload of a secret is flexible
- Secrets have been a stable part of the Kubernetes ecosystem since Kubernetes 1.2

Because of these features, the Brigade system uses secrets for bearing event information.

### The Anatomy of a Brigade Event Secret

Here is the structure of a Brigade event secret. It is annotated to explain what
data belongs in what fields.

```yaml
# The main structure is a normal Kubernetes secret
apiVersion: v1
kind: Secret
metadata:
  # Every event has an automatically generated name. The main requirement of
  # this is that it MUST BE UNIQUE.
  name: example
  # Brigade uses several labels to determine whether a secret carries a
  # Brigade event.
  labels:
    # 'heritage: brigade' is mandatory, and signals that this is a Brigade event.
    heritage: brigade

    # This should point to the Brigade project ID in which this event is to be
    # executed
    project: brigade-1234567890

    # This MUST be a unique ID. Where possible, it SHOULD be a ULID
    # Substituting a UUID is fine, though some sorting functions won't be as
    # expected. (A UUID v1 will be sortable like ULIDs, but longer).
    build: 01C1R2SYTYAR2WQ2DKNTW8SH08

    # 'component: build' is REQUIRED and tells brigade to create a new build
    # record (and trigger a new worker run).
    component: build

    # Any other labels you add will be ignored by Brigade.
type: brigade.sh/build
data:
  # IMPORTANT: We show these fields as clear text, but they MUST be base-64
  # encoded.

  # The name of the thing that caused this event.
  event_provider: github

  # The type of event. This field is freeform. Brigade does not have a list of
  # pre-approved event names. Thus, you can define your own event_type
  event_type: push

  # Revision describes a vcs revision.
  revision:

    # Commit is the commitish/reference for any associated VCS repository. By
    # default, this should be `master` for Git.
    commit: 6913b2703df943fed7a135b671f3efdafd92dbf3

    # Ref is the symbolic ref name. (refs/heads/master, refs/pull/12/head, refs/tags/v0.1.0)
    ref: master

  # This should be the same as the `name` field on the secret
  build_name: example

  # This should be the same as the 'project' label
  project_id: brigade-1234567890

  # This should be the same as the 'build' label
  build_id: 01C1R2SYTYAR2WQ2DKNTW8SH08

  # The payload can contain arbitrary data that will be passed to the worker
  # JavaScript. It is passed to the script unparsed, and the script can parse
  # it as desired.
  payload: "{ 'foo': 'bar' }"

  # An event can supply a script to execute. If it does not supply a script,
  # Brigade will try to locate a 'brigade.js' file in the project's source
  # code repository using the commit provided above.
  script: "console.log('hello');"
```

Again, note that any fields in the `data:` section above are shown cleartext,
though in reality you _must_ base-64 encode them.

The easiest way to create a secret like the above is to do so with the `kubectl`
command, though there are a host of language-specific libraries now for creating
secrets in code.

## Creating Custom Gateways

Given the above description of how gateways work, we can now talk about a gateway
as anything that generates a secret following the format above.

In this final section, we will create a simple shell script that triggers a new
event every 60 seconds. In the payload, it sends the system time of the host
that is running the script.

```bash
#!/usr/bin/env bash
set -euo pipefail

# The Kubernetes namespace in which Brigade is running.
namespace="default"

event_provider="simple-event"
event_type="my_event"

# This is github.com/brigadecore/empty-testbed
project_id="brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac"
commit_ref="master"
commit_id="589e15029e1e44dee48de4800daf1f78e64287c0"

base64=(base64)
uuidgen=(uuidgen)
if [[ "$(uname)" != "Darwin" ]]; then
  base64+=(-w 0)
  uuidgen+=(-t) # generate UUID v1 for sortability
fi

# This is the brigade script to execute
script=$(cat <<EOF
const { events } = require("brigadier");
events.on("my_event", (e) => {
  console.log("The system time is " + e.payload);
});
EOF
)

# Now we will generate a new event every 60 seconds.
while :; do
  # We'll use a UUID instead of a ULID. But if you want a ULID generator, you
  # can grab one here: https://github.com/technosophos/ulid
  uuid="$("${uuidgen[@]}" | tr '[:upper:]' '[:lower:]')"

  # We can use the UUID to make sure we get a unique name
  name="simple-event-$uuid"

  # This will just print the system time for the system running the script.
  payload=$(date)

  cat <<EOF | kubectl --namespace ${namespace} create -f -
  apiVersion: v1
  kind: Secret
  metadata:
    name: ${name}
    labels:
      heritage: brigade
      project: ${project_id}
      build: ${uuid}
      component: build
  type: "brigade.sh/build"
  data:
    revision:
      commit: $("${base64[@]}" <<<"${commit_id}")
      ref: $("${base64[@]}" <<<"${commit_ref}")
    event_provider: $("${base64[@]}" <<<"${event_provider}")
    event_type: $("${base64[@]}" <<<"${event_type}")
    project_id: $("${base64[@]}" <<<"${project_id}")
    build_id: $("${base64[@]}" <<<"${uuid}")
    payload: $("${base64[@]}" <<<"${payload}")
    script: $("${base64[@]}" <<<"${script}")
EOF
  sleep 60
done
```

While the main point of the script above is just to show how to create a basic
event, it should also demonstrate how flexible the system is. A script can take
input from just about anything and use it to trigger a new event.

## Creating A Cron Job Gateway

Beginning with the code above, we can build a gateway that runs as a scheduled
job in Kubernetes. In this example, we use a Kubernetes
[CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/) object
to create the secret.

First we can begin with a simplified version of the script above. This one does not
run in a loop. It just runs once to completion.

Here is `cron-event.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

# The Kubernetes namespace in which Brigade is running.
namespace=${NAMESPACE:-default}

event_provider="simple-event"
event_type="my_event"
project_id="brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac"
commit_ref="master"
commit_id="589e15029e1e44dee48de4800daf1f78e64287c0"
uuid="$(uuidgen | tr '[:upper:]' '[:lower:]')"
name="simple-event-$uuid"

payload=$(date)
script=$(cat <<EOF
const { events } = require("brigadier");
events.on("my_event", (e) => {
  console.log("The system time is " + e.payload);
});
EOF
)

cat <<EOF | kubectl --namespace ${namespace} create -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${name}
  labels:
    heritage: brigade
    project: ${project_id}
    build: ${uuid}
    component: build
type: "brigade.sh/build"
data:
  revision:
    commit: $(base64 -w 0 <<<"${commit_id}")
    ref: $(base64 -w 0 <<<"${commit_ref}")
  event_provider: $(base64 -w 0 <<<"${event_provider}")
  event_type: $(base64 -w 0 <<<"${event_type}")
  project_id: $(base64 -w 0 <<<"${project_id}")
  build_id: $(base64 -w 0 <<<"${uuid}")
  payload: $(base64 -w 0 <<<"${payload}")
  script: $(base64 -w 0 <<<"${script}")
EOF
```

Next, we will package the above as a Docker image. To do that, we create a `Dockerfile`
in the same directory as the `cron-event.sh` script above.

The `Dockerfile` just sets up the commands we need, and then copies the script into
the image:

```Dockerfile
FROM debian:jessie-slim
RUN apt-get update && apt-get install -y uuid-runtime curl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
  && mv kubectl /usr/local/bin/kubectl && chmod 755 /usr/local/bin/kubectl
COPY ./cron-event.sh /usr/local/bin/cron-event.sh
CMD /usr/local/bin/cron-event.sh
```

(The really long line just installs `kubectl`)

And we can pack that into a Docker image by running `docker build -t technosophos/example-cron:latest .`.
You should replace `technosophos` with your Dockerhub username (or modify the
above to store in your Docker registry of choice).

Then push the image to a repository that your Kubernetes cluster can reach:

```
$ docker push technosophos/example-cron
```

Now we create our third (and last) file: a CronJob definition. Our `cron.yaml`
should look something like this:

```yaml
apiVersion: batch/v1beta1
kind: CronJob
metadata:
  name: example-cron-gateway
  labels:
    heritage: brigade
    component: gateway
spec:
  schedule: "*/1 * * * *"
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
            - name: cron-example
              image: technosophos/example-cron:latest
              imagePullPolicy: IfNotPresent
```

We can install it with `kubectl create -f cron.yaml`. Now, every minute our
new gateway will create an event.

Whenever you are done with this example, you can delete it with
`kubectl delete cronjob example-cron-gateway`.

That's it! We have create both a local shell script gateway and an in-cluster
cron gateway.

Again, there are other programming libraries and platforms that interoperate with
Kubernetes. Many of them are hosted on GitHub in the [kubernetes-client org](https://github.com/kubernetes-client).
