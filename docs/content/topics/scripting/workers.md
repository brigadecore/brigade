---
title: Workers
description: How to use a custom Worker in Brigade
section: scripting
weight: 5
aliases:
  - /workers.md
  - /topics/workers.md
  - /topics/scripting/workers.md
---

# What is a Brigade Worker?

A worker is the Brigade component that is launched in response to an event that
a project subscribes to. There is a one-to-one relationship between events and
workers.

_Typically_, the action of a worker is driven by the contents of a project's
`brigade.js` (or `brigade.ts`) script, however, _custom_ workers aren't
strictly limited to this approach and can utilize alternative mechanisms for
defining project/event-specific configuration or logic.

The remainder of this page covers the general approach of creating and using
custom workers.

# Creating a Custom Worker

Although it is possible to extend the default Brigade Worker with additional
Node.js libraries and/or custom JavaScript or TypeScript code (indeed, we've
seen how this can be achieved via Brigade's [Dependencies] model), nothing
prevents you from defining a new class of worker from scratch.

This approach allows significantly more flexibility with respect to how the
worker is implemented (languages, frameworks, etc.) _and_ how the worker
functions. For instance, it is entirely possible to create workers that bypass
`brigade.js` and drive builds based on some other declarative or imperative
format.

The remainder of this section covers the general requirements for a custom
worker as well as the methods whereby Brigade passes configuration to a worker
for processing an event.

[Dependencies]: /topics/scripting/dependencies

## Event Data

Workers get the data they need to process an event via JSON mounted to its pod
by Brigade, each time it runs. This JSON contains configuration that the worker
itself uses, as well as project- and event-specific details. This file is
located at `/var/event/event.json` inside the worker container.

Note that custom workers may selectively disregard any data they deem
inapplicable to the custom behavior they implement. For instance, a field that
conveys the expected location of the `brigade.js` file is not applicable to a
worker that does not utilize such configuration.

### Fields

Here's a rundown on the data present in the event JSON:

| Field Name | Description |
|------------|-------------|
| `id` | A unique ID for the event |
| `project.id` | The project ID that this event is associated with |
| `project.kubernetes.namespace` | The Kubernetes namespace that the project uses |
| `project.secrets` | The key/value map of [secrets] for the project |
| `source` | The event [source] |
| `type` | The event [type] |
| `qualifiers` | The event [qualifiers] |
| `labels` | The event [labels] |
| `shortTitle` | A [short title] for the event |
| `longTitle` | A [long title] for the event |
| `payload` | The event [payload] |
| `worker.apiAddress` | The address for the Brigade API server |
| `worker.apiToken` | The token used for communicating with the Brigade API server |
| `worker.logLevel` | The log level for worker logs |
| `worker.configFilesDirectory` | The directory in the worker where the Brigade config files may be found |
| `worker.defaultConfigFiles` | The default config files to use if not provided elsewhere (e.g. via a git repository's source code) |
| `worker.git.cloneURL` | The clone URL of the git repository associated with the worker |
| `worker.git.commit` | The commit SHA for the git repository |
| `worker.git.ref` | The reference for the git repository |
| `worker.git.initSubmodules` | Whether or not git submodules should be initialized |

[secrets]: /topics/project-developers/secrets
[source]: /topics/project-developers/events#source
[type]: /topics/project-developers/events#type
[qualifiers]: /topics/project-developers/events#qualifiers
[labels]: /topics/project-developers/events#labels
[short title]: /topics/project-developers/events#short-title
[long title]: /topics/project-developers/events#long-title
[payload]: /topics/project-developers/events#payload

## Job Pod Names and Labels

To be visible to the rest of Brigade, certain naming and labeling conventions
must be adhered to for job pods spawned by the worker.

1. Pod names MUST take the form `<event ID>-<job name>`.

1. Pods MUST be labeled as follows:

    | Key | Value |
    |-----|-------|
    | `brigade.sh/component` | `job` |
    | `brigade.sh/job` | Job Name |
    | `brigade.sh/event` | Event ID |
    | `brigade.sh/id` | Brigade's ID<sup>*</sup> |
    | `brigade.sh/project` | Project ID |

<sup>*</sup> Brigade's ID is based on the Kubernetes namespace that Brigade is
deployed to and the deployed release name of Brigade, e.g.
`brigade-namespace.brigade-release`

All the [usual constraints][K8s label constraints] in labeling Kubernetes pods
apply.

[K8s label constraints]: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set

## Exit Code

Brigade determines the ultimate success or failure of an event by the return
code from the worker.

Workers that successfully execute to completion with no errors MUST exit with
return code 0.

Worker executions that fail MUST exit with a non-zero return code.

# Building and Publishing a Custom Worker Image

When the custom worker code is ready to be used in Brigade, the next step is to
build a Docker image from your `Dockerfile`.

```console
$ docker build -t myregistry/myworker:latest .
$ docker push myregistry/myworker:latest
```

> IMPORTANT: Make sure you replace `myregistry` and `myworker` with your own
> account and image names.

**Tip:** If you are running a local Kubernetes cluster with Docker or Minikube,
you do not need to push the image. Just configure your Docker client
to point to the same Docker daemon that your Kubernetes cluster is using. (With
Minikube, you do this by running `eval $(minikube docker-env)`.)

Now that we have our image pushed to a usable location, we can configure Brigade
to use this new image.

# Configuring Brigade to Use Your Custom Worker Image

The worker image for Brigade can be set globally as well as individually per
project.

## Global default

To set the version globally, you should supply values for the `repository`,
`tag` and, optionally, `pullPolicy` fields under the `worker.image` section in
the [Brigade chart].

Using the same example values from above, this would look like:

```yaml
worker:
  image:
    repository: myregistry/myworker
    tag: latest
    # The default is 'IfNotPresent', but you may wish to set to 'Always' when a
    # mutable tag is used
    pullPolicy: Always
```

You can then use Helm's [upgrade command] to update Brigade with these new
values.

[Brigade chart]: https://github.com/brigadecore/brigade/tree/v2/charts/brigade
[upgrade command]: https://helm.sh/docs/helm/helm_upgrade/
## Project Overrides

To configure the worker image per-project, you'll supply the same repository,
tag and pullPolicy values to the project definition under the
`spec.workerTemplate` section:

```yaml
spec:
  workerTemplate:
    container:
      image: myregistry/myworker:latest
      imagePullPolicy: Always
```

Then, update the project via `brig project update --file project.yaml`

## Using Your Custom Worker

Once you have Brigade and/or individual projects updated, your new Brigade
workers will automatically switch to using this new image.
