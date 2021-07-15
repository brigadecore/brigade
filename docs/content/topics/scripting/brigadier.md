---
title: The Brigade.js API
description: Describing the public APIs typically used for writing Brigade.js
aliases:
  - /brigadier.md
  - /topics/brigadier.md
  - /topics/scripting/brigadier.md
---

TODO: update per v2

# The Brigade.js API

This document describes the public APIs typically used for writing Brigade.js. It does not
describe internal libraries, nor does it list non-public methods and properties on
these objects.

An Brigade JavaScript file is executed inside of a cluster. It runs inside of a
Node.js-like environment (with a few libraries blocked for security reasons). It
uses Node 8.

## High-level Concepts

An Brigade JS file is always associated with a _project_. A project defines contextual
information, and also dictates the security parameters under which the script will
execute.

A project may associate the script to a _repository_, where a repository is typically
a VCS reference (e.g. a git repository). Each job will, by default, have access
to the project's repository.

Brigade files respond to _events_. That is, Brigade scripts are typically composed of one or
more _event handlers_. When the Brigade environment triggers an event, the associated
event handler will be called.

## The `brigadier` Library

The main library for Brigade is called `brigadier`. The Brigade runtime grants access to
this library.  The source code for this library is located in [brigadecore/brigadier](https://github.com/brigadecore/brigadier).

```
const brigadier = require('brigadier')
```

It is considered idiomatic to destructure the library on import:

```
const { events, Job, Group } = require('brigadier')
```

Some objects described in this document are not declared in `brigadier`, but are
exposed via `brigadier`.

### The `BrigadeEvent` class

The `BrigadeEvent` class describes an event. Typically, it is exposed to the script
via a callback handler.

```
events.on("pull", (brigadeEvent, project) => {})
```

An instance of an `BrigadeEvent` has the following properties:

- `buildID: string`: The unique ID for the build. This will change for each build.
- `type: string`: The event type (`push`, `exec`, `pull_request`).
- `provider: string`: The name of the thing that triggered this event.
- `revision: Revision`: The revision details, if supplied, of the underlying VCS system.
- `payload: string`: Arbitrary data supplied by an event emitter. Each event emitter
  will describe its own payload. For example, the GitHub gateway emits events that
  contain GitHub's webhook objects.
- `cause: Cause`: If one event triggers another event, the causal chain is passed
  through the `cause` property

### The `revision` object

The `revision` object has the following properties:

- `commit: string`: The commit ID, if supplied, for the underlying VCS system. When this is supplied, each Job will have access to the VCS at this revision.
- `ref: string`: The symbolic ref name. (e.g `refs/heads/master`)

If the `revision` object is not provided, it may be interpreted as `master`, or the head of the main branch.
_The default value is not guaranteed to be `master` in future versions._

#### The `Cause` class

A `Cause` is attached to an `BrigadeEvent`, and describes the event that caused this
event. It has the following properties:

- `event: BrigadeEvent`: The causing event
- `reason: any`: The reason this event was caused. Typically this is an error object.
- `trigger: string`: The mechanism that triggered this event (e.g. "unhandled exception")

The `after` and `error` built-in events will set a `Cause` on their `BrigadeEvent` objects.

### The `events` Object

Within `brigadier`, the `events` object provides access to the main event handler.

#### `events.on(eventName: string, callback: (e: BrigadeEvent, p: Project) => {})`

The `events.on()` function is the way event handlers are registered. An `on()` method
takes two arguments: the name of the event and the callback that will be executed
when the named event fires.

```javascript
events.on("push", (e, p) => {
  console.log(p.name);
});
```

#### `events.has(eventName: string): boolean`

`events.has` is used to see if an event handler was registered already.

### The `Group` class

The `Group` class provides both static methods and object methods for working
with groups.

#### The static `runAll(Job[]): Promise<Result[]>` method

The `runAll` method runs all jobs in parallel, and returns a Promise that waits until
all jobs are done and then returns the collected results.

This is useful for running a batch of jobs in parallel, but waiting until they are
complete before continuing with another operation.

#### The static `runEach(Job[]): Promise<Result[]>` method

This runs each of the given jobs in sequence, blocking on each job until it
is complete. The Promise will return the collected results.

#### The `new Group(Job[]): Group` constructor

Create a new `Group` and optionally pass it some jobs.

#### The `add(Job...)` method

Adds one or more Job objects to the group.

#### The `length(): number` method

Return how many jobs are in the group.

#### The `runAll(): Promise<Result[]>` method

Runs all of the jobs in the group in parallel. When the Promise resolves, it will
wrap all of the results.

Functionally, this is equivalent to the static `runAll` method.

#### The `runEach` method

Runs each of the jobs in sequence (synchronously). When the Promise resolves, it will
wrap all of the results.

Functionally, this is equivalent to the static `runEach` method.

### The `Job` class

The `Job` class describes a job that can be run.

#### constructor `new Job(name: string, image?: string, tasks?: string[], imageForcePull?: boolean): Job`

The constructor requires a `name` parameter, and this must be unique within your
script. It must be composed of the characters a-z, A-Z, 0-9, and `-`. Additionally,
the `-` cannot be the first or last character, and the name must be at least two
characters.

Optionally, you may specify the container image (e.g. `node:8`, `alpine:3.4`). The
container image must be fetchable by the runtime (Kubernetes). If no container is
specified here or with `Job.image`, a default image will be loaded.

Optionally, you may specify a list of tasks to be run inside of the container. If no
tasks are specified here or with `Job.tasks`, the container will be run with its
defaults.

These two are equivalent:

```javascript
var one = new Job("one");
one.image = "alpine:3.4";
one.tasks = ["echo hello"];

var two = new Job("two", "alpine:3.4", ["echo hello"]);
```

Properties of `Job`

- `name: string`: The job name
- `shell: string`: The shell in which to execute the tasks (`/bin/sh`)
- `tasks: string[]`: Tasks to be run in the job, in order. Tasks are concatenated
  together and, by default, packaged as a Bourne (`/bin/sh`) shell script with `set -e`.
  If the Bourne Again Shell is used (`/bin/bash`), `set -eo pipefail` will be used.
- `args: string[]`: Arguments to pass to the container's entrypoint. It is recommended,
  though not required, that implementors not use both `args` and `tasks`.
- `imageForcePull: boolean`: Defines the container image pull policy: `Always` if `true` or `IfNotPresent` if `false` (defaults to `false`).
- `env: {[key: string]:string}`: Name/value pairs of environment variables.
- `image: string`: The container image to run
- `imagePullSecrets: string[]`: The names of the pull secrets (for pulling images from a secure remote repository)
- `mountPath: string`: The path where any resources should be mounted (e.g. where a Git repository will be cloned) (defaults to `/src`)
- `timeout: number`: Time to wait, in milliseconds, before the job is marked "failed"
- `useSource: bool`: If false, no external resource will be loaded (e.g. no git clone will be performed)
- `privileged: bool`: If this is true, the job will be executed in privileged mode, which allows it to do things like access a Docker socket. EXPERTS ONLY.
- `host: JobHost`: Preferences for the host that runs the job.
- `cache: JobCache`: Preferences for the job's cache
- `storage: JobStorage`: Preferences for the way this job attaches to the build storage
- `docker: JobDockerMount`: Preferences for mounting a Docker socket
- `serviceAccount: string`: The name of the service account to use (if you need to override the default).
- `annotations: {[key: string]:string}`: Name/value pairs of annotations to add to the job's pod
- `resourceRequests: JobResourceRequest`: CPU and memory request resources for the job pod container.
- `resourceLimits: JobResourceLimit`: CPU and memory limit resources for the job pod container.
- `streamLogs: boolean`: controls whether logs from the job Pod will be streamed to output (similar functionality to `kubectl logs PODNAME -f`).
- `volumes: kubernetes.V1Volume[]`: list of Kubernetes volumes to be attached to the job pod specification. See the [Kubernetes type definition](https://github.com/kubernetes-client/javascript/blob/159b32d2cd96117eb19342190c6bc3fa9bc8e3eb/src/gen/model/v1Volume.ts#L44)
- `volumeMounts: kubernetes.V1VolumeMount[]`: list of Kubernetes volume mounts to be attached to all containers in the job pod specification. See the [Kubernetes type definition](https://github.com/kubernetes-client/javascript/blob/159b32d2cd96117eb19342190c6bc3fa9bc8e3eb/src/gen/model/v1VolumeMount.ts#L17)
#### Setting execution resources to a job

For some jobs is a good practice to set limits and guarantee some resources. In the following example job pod container resource requests and limits are set.

```javascript
var job = new Job("huge-job");

// Our job uses a lot of resources, we set huge requests but set safe memory limits:
job.resourceRequests.memory = "2Gi";
job.resourceRequests.cpu = "500m";
job.resourceLimits.memory = "3Gi";
job.resourceLimits.cpu = "1";
```

All are optional, for example you could set only `resourceLimits.memory = 3Gi`).

#### The `job.podName()` method

This returns the name of the pod that was started during `job.run()`. It will return
an empty string before `run()` is called.

#### The `job.run(): Promise<Result>` method

Run the job, returning a Promise that returns when the job is complete.

### The `JobCache` class

A `JobCache` object provides preferences for a job's usage of a cache.

Caches are disabled by default.

Properties:

- `enabled: boolean`: If `true`, the cache is turned on for this job.
- `size: string`: The size, defaults to `5Mi`. This value is only evaluated the first
  time a job is cached. To resize, the cache must be destroyed manually.
- `path: string`: A read-only attribute returning path (in the container) in which the cache
  is available.

### The `JobDockerMount` class

The `JobDockerMount` controls whether, and how, a Docker socket is mounted to
the job. Docker sockets are used for building Docker images. Because they mount
to the host, using a Docker socket is considered dangerous. Thus, to use the
Docker mount, the job must be put into `privileged` mode.

Properties:

- `enabled: boolean`: If `true`, the Docker socket will be mounted to the pod

### The `JobHost` class

A `JobHost` object provides preferences for the host upon which the job is executed.

- `os: string`: The name of the OS upon which the job should be run (`linux`, `windows`).
  Not all clusters support all OSes.
- `name: string`: The name of the host (node) upon which the job will run. This is
  highly system dependent.
- `nodeSelector: Map<string, string>`: Labels which are used as selectors when choosing a node on which to run this job.
In the following example job nodeSelector is set.  
```javascript
var job = new Job("job-on-specific-group-of-nodes");

// We want to schedule job pod on a specific group of nodes:
job.host.nodeSelector.set('nodegroup', 'brigade-node');
```

### The `JobStorage` class

- `enabled: boolean`: If set to `true`, the Job will mount the build storage.
  Build storage exposes a mounted volume at `/mnt/brigade/share` with storage that
  can be shared across jobs.
- `path: string`: The read-only path to the shared storage from within the container.

### The `KubernetesConfig` class

A KubernetesConfig object has the following properties:

- `namespace: string`: The namespace in which Kubernetes objects are created.
- `vcsSidecar: string`: The name of the sidecar image that fetches the repository.
  By default, this is the Git sidecar that fetches git repositories.
- `buildStorageSize: string`: The size of the build shared storage space used by the build jobs.

### The `Result` class

This wraps the result of a Job run.

#### The `toString(): string` method

This returns the result as a string.

### The `Project` class

Properties:

- `id: string`: The unique ID of the project
- `name: string`: The project name, typically `org/name`.
- `kubernetes: KubernetesConfig`: The object describing this project's Kubernetes settings
- `repo: Repository`: Information on the upstream repository (if available).
- `secrets: {[key: string]: string}`: Key/value pairs of secret name and secret value.
  The security model _may_ limit access to this property or its values.

Secrets (`project.secrets`) are passed from the project configuration into a Kubernetes Secret, then injected into Brigade.

So `helm install brigade-project --set secrets.foo=bar` will add `foo: bar` to
`project.secrets`.

### The `Event` object

The Event object describes an event.

Properties:

- `type`: The event type (e.g. `push`)
- `provider`: The entity that caused the event (`github`)
- `revision`: The [Revision](#the-revision-object) object containing details for the commit that this script should operate on.
- `payload`: The object received from the event trigger. For GitHub requests, its
  the data we get from GitHub.

### The `Job` object

To create a new job:

```javascript
j = new Job(name);
```

Parameters:

- A job name (alpha-numeric characters plus dashes).

Properties:

- `name`: The name of the job
- `image`: A Docker image with optional tag.
- `tasks`: An array of commands to run for this job
- `shell`: The terminal emulator that job tasks will be executed under. By default,
  this is /bin/sh
- `env`: Key/value pairs or Kubernetes value references that will be injected into the environment.
  - If supplying key/value, the key is the variable name (`MY_VAR`), and the value is the string value (`foo`)
  - If you are referencing existing Secrets or ConfigMaps in your Kubernetes cluster, the `env` object key
    will be your secret name, and the value will be a Kubernetes reference object. `fieldRef`, `secretKeyRef`,
    and `configMapKeyRef` are accepted. `resourceFieldRef` is technically supported but not advised, since resources
    are not generally specified for Brigade jobs.
  - Example:
    ```javascript
    myJob.env = {
      myOneOffSecret: "secret value",
      myConfigReference: {
        configMapKeyRef: {
          name: "my-configmap",
          key: "my-configmap-key"
        }
      },
      mySecretReference: {
        secretKeyRef: {
          name: "my-secret",
          key: "my-secret-key"
        }
      }
    };
    ```

It is common to pass data from the `e.env` Event object into the Job object as is appropriate:

```javascript
events.push = function(e) {
  j = new Job("example");
  j.env = { DB_PASSWORD: project.secrets.dbPassword };
  //...
  j.run();
};
```

The above will make `$DB_PASSWORD` available to the "example" job's runtime.

Methods:

- `run()`: Run this job and wait for it to exit.
- `background()`: Run this job in the background.
- `wait()`: Wait for a backgrounded job to complete.

### The `Repository` Class

The `Repository` class describes a project's VCS repository (if provided).

- `name: string`: The name of the repo (`org/name`)
- `cloneURL: string`: The URL that the VCS software can use to clone the repository.
