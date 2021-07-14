---
title: Design
description: How Brigade is designed
aliases:
  - /design.md
  - /topics/design.md
---

TODO: update per v2

# Brigade Design

_This is a living document, and is kept up to date with the current state of
Brigade. It is a high-level explanation of the Brigade design._

Brigade is an in-cluster runtime environment. It interprets scripts, and executes
them often by invoking resources inside of the cluster. Brigade is event-based
scripting of Kubernetes pipelines.

![Event-based scripting of pipelines](https://docs.brigade.sh/img/design-01.png)

- Event-based: A script execution is triggered by a Brigade event.
- Scripting: Programs are expressed as JavaScript files that declare one or more event handlers.
- Pipelines: A script expresses a series of related jobs that run in parallel or serially.

## Terminology

![Brigade Run](https://docs.brigade.sh/img/design-02.png)

- **Brigade** is the name of the project. Often, the term is used generically to
  refer to the in-cluster Brigade components.
  - **Brigade Controller** is the name of the central controller.
  - **Brigade API** is an API server used to access information about Brigade's
    current and past workloads.
- **Brig** is a command line client for Brigade.
- **Event** is a Brigade-issued indicator that something happened.
- **Gateways** transform outside triggers (a Git pull request, a Trello card move) into
  events.
- **brigade.js** is the standard name for a JavaScript file that contains one or more Brigade event handlers.
- **Job** is a build unit, comprised of one or more build steps called "tasks"
  - **Task** is a step within a job.
- **Webhook** is an incoming HTTP/HTTPS request from an external source. Some gateways
  uses webhooks as triggers for events.
- **Project** is a named collection of data that Brigade operates on. Often, though
  not always, a Project is linked to a Git project. All scripts are executed within
  the context of a project.
  - **Secrets** are bits of configuration considered "non-public". They are stored
  in projects, and may be accessed within scripts.
  - **VCS Sidecar** is a special Docker image that knows how to load a project's
  related VCS repository into a build or job. The default VCS sidecar only knows
  about Git.
- **Build** is a run (an instance) of a script. When a script is executed, the
  data about that execution is conceptually referenced as a _build_. A build
  must handle at least one event, and may handle multiple events.

## The Developer's View

From the developer's view, Brigade works like this:

A _project_ describes the context in which a Brigade script will run. It may define
the following:

- Project name
- Related files, stored in a VCS repository (Git is supported out of the box, others
  require a different VCS sidecar)
- Configuration, such as tokens, SSH keys, and other items necessary to wire up
  a script run.
- Secrets, which are opaque configuration items that are presumed to be non-public.
  While configuration information is not guaranteed to be available inside of a
  script, secrets are.
- The VCS sidecar to use. (Default is the Git sidecar)

One or more scripts may be executed within the context of a project. Brigade
assumes that a default script will reside in the project's VCS repository at the
relative path `./brigade.js`. Gateways may provide other ways of sending
scripts into Brigade.

A Brigade script should have at least one event handler defined. Event handlers
are triggered when a gateway emits an event. Events are bound to projects, so
an event will only be triggered for the explicitly declared project. (In other
words, there are no _global_ events, only project-bound events.)

An event specifies the following things:

- Which project it is attached to (e.g. `my/project`)
- The name of the event that fired (e.g. `pull_request`)
- The entity that triggered the event (e.g. `github`)
- The script to run (defaults to `./brigade.js` in the VCS)
- A payload containing event data.

(The list above is not exhaustive.)

When Brigade receives an event, it loads the referenced project, then starts a
new worker. The worker executes the Brigade script, using as its entry point the
event handler for the triggered event (e.g. `events.on('pull_request')`. The worker
processes the script until one of the following occurs:

- The script exits successfully
- The script terminates because of an error
- The worker times out

The fundamental units of a script are event handlers, jobs, and tasks.

An _event handler_ associates a named event with a function that can process the
event:

```javascript
events.on('event name', () => { /* handler */ })
```

An event handler is explicitly given two pieces of information: the `event` record
and the `project` record.

- The `event` record provides information about the event.
- The `project` record provides some (but not all) information about the project,
  notably names, secrets, and VCS information

A typical event handler declares and runs one or more _jobs_. A job is a discrete
unit of work that is associated with a _container image_.

```javascript
const myJob = new Job("job-name", "image:tag")
```

When a job is executed, the container image is _pulled_ from an origin (such as DockerHub),
and is executed in the cluster. A job specifies configuration and input to that
container. And the output of that container is returned from the job.

A job may declare zero or more _tasks_. A task is an individual step executed inside
of a container. For example, if a container is just a simple Linux container with
a shell, multiple shell commands can be run as tasks:

```
myJob.tasks = [
  "echo hello",
  "echo world"
]
```

In addition to jobs, scripts may declare _groups_, where groups are merely organizing
units that can execute multiple jobs according to predefined patterns (e.g. all
in parallel, each serially).

When a script is executed, cluster resources are allocated to execute each job as
an independent cluster resource (a Pod). Various storage configurations may provide
shared space between jobs in a build, or between multiple instances of the same
job in different builds.

## The Operator's View

Operationally speaking, Brigade is a tool for chaining together Kubernetes pods
in order to accomplish high level goals. It is analogous to the way UNIX shell scripts
work.

A UNIX shell script defines the workflow around executing one or more lower-level
system executables. Similarly, a Brigade script defines a workflow for executing
multiple containers within a cluster.

Brigade has several functional concepts.

![Design Overview](https://docs.brigade.sh/img/design-overview.png)

A Gateway is a workload, typically a Kubernetes Deployment fronted by a Service
or Ingress, that transforms a trigger (inbound webhook, item on queue) into a
Brigade event.

![Service, Trigger, Gateway, Event](https://docs.brigade.sh/img/design-trigger-gateway.png)

The illustration above shows how GitHub translates a Git event into a webhook, which
the optional [Brigade GitHub Gateway](./github.md) translates into an event to be
consumed by the Brigade controller.

In Brigade, all scripts are executed in the context of a _project_. Projects are
represented as Kubernetes Secrets.

The Controller is a Kubernetes controller that listens for Brigade event objects,
and handles these objects by starting workers.

Brigade events are currently specified as Kubernetes Secrets with particular
labels. We use secrets because at the time of development, Third Party Resources
were deprecated and Custom Resource Descriptions are not final.

Brigade Workers are pods that execute brigade scripts. Each worker handles exactly
one brigade script. Workers are never pooled. A worker runs to completion, to failure,
or to timeout.

Brigade workers handle an event by starting a _build_, where a build executes a
script. A build will create a PVC for shared storage (job-to-job shared filesystem),
and will create one or more pods (one per job). The worker will attempt to destroy
all destroyable resources once the build has completed. Note that jobs are left in
the Complete state (not deleted) so that their logs may be accessed. Cache PVCs
are left unattached, and prepared for re-use.

A Brigade Job is started by a worker, and is executed as a pod. A job is run
to completion, to error, or to timeout. Its status and results are made available
to the calling worker, which in turn provides access to the script.

Along with the execution of an event-build pipeline, Brigade also provides an
API server that provides access to information about current and past builds, projects,
and jobs. The API server is typically fronted by a Service or Ingress.

## Reasoning for Certain Design Decisions

- Go was selected because it provides the most mature Kubernetes APIs.
- At various points, we explored using TPRs and later CRDs. But the feature sets
  and stability of these two facilities never reached a satisfactorily stable
  point. So we use secrets for configuration data. Secrets have the following benefits:
  - They are mountable via the volume system, a feature we use frequently
  - They can be injected as environment variables
  - They benefit from Kubernetes improvements to secret storage
- JavaScript was selected because of it's high score on just about all language
  usage analyses.
- Node.js was selected because of its robust ecosystem
- TypeScript was selected because its type system resembled Go's.
- The Controller model was selected because it gave us the advantages of a queueing
  system, but without requiring a stand-alone service.
- GitHub and Git were selected because they appear to be the leading tools used
  by developers.
- Docker containers were selected because they are the clear market leader.
- PVCs are used for shared storage because they are the closest Kubernetes comes
  to providing a shareable filesystem. We explored, and ultimately rejected, multiple
  userland alternatives.
- The Job/Task terminology comes from ETL
- The Event terminology comes from JavaScript's event model
- The Build terminology comes from CI/CD systems

## History of Brigade

Brigade was designed in March 2016 by the Deis Helm Team (now part of Microsoft).

The first design used Lua instead of JavaScript, and relied on very few Kubernetes resources.
Instead, it used a Redis queue for message passing and key/value storage. Other
than some proof-of-concept work, the Lua engine never materialized. JavaScript's
popularity made it a better choice.

An original Kubernetes-oriented JavaScript engine was developed several months later.
This was intended to be both a stand-alone component and a foundational piece for
Brigade. Work was abandoned in favor of the Node.js worker pattern.

In April of 2017, Brigade was designated as the third part of the Helm/Draft/Brigade
ecosystem. At this point, it was renamed "Acid" (Acme Continuous Integration & Deployment).

Brigade reached a stability point in September 2017, and was re-renamed back from
Acid to Brigade. Brigade was released publicly under the MIT license in October
2017, as release 0.1.0.

In March of 2019, in tandem with Brigade's 1.0.0 release, the project was donated to
the CNCF org as a Sandbox Project.  During this transition, the license was updated to
Apache-2.0.
