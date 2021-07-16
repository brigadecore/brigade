---
title: 'Workers'
description: 'How to add custom libraries to a Brigade worker'
aliases:
  - /workers.md
  - /topics/workers.md
  - /topics/scripting/workers.md
---

TODO: update per v2

# What is a Brigade Worker?

A worker is a Brigade component that the Brigade controller launches in response
to an event in order to execute a build in accordance with
project/event-specific configuration or logic. There is a one-to-one
relationship between events and workers.

_Typically_, the action of a worker is driven by the contents of project's
`brigade.js` file, however, _custom_ workers aren't strictly limited to this
approach and can utilize alternative mechanisms for defining
project/event-specific configuration or logic.

The remainder of this page covers various methods of creating and using
custom workers.

# Extending the Default Worker with Additional NPM Packages

Brigade ships with a worker focused on running jobs defined by a project's
`brigade.js` file. This worker exposes a host of useful Node.js libraries to
`brigade.js` files, as well as the `brigadier` Brigade library.

Sometimes, it is necessary to make additional libraries (perhaps even
custom libraries) available to your `brigade.js` file.  There are two methods
available to achieve this:

1. Supplying a `brigade.json` file listing dependencies. Refer to [the dependencies
  document](dependencies.md) for a detailed description of this approach.

1. Create a custom Docker image for the worker that contains the additional
   dependencies.

The remainder of this section focuses on the second approach.

Since the Brigade worker (`brigade-worker`) is supplied as a Docker image,
amending the default worker with additional Node.js libraries is as simple as
"extending" the default worker's Docker image to create a custom image.

By way of example, suppose we wish to provide our `brigade.js` with access to
an XML parser library. This can be accomplished using the following
`Dockerfile`:

```Dockerfile
FROM brigadecore/brigade-worker:v1.2.1

RUN yarn add xml-simple
```

The `Dockerfile` begins with the default Brigade worker image and simply adds
the `xml-simple` library.

Next, skip to the section on [building and publishing a custom worker image](#building-and-publishing-a-custom-worker-image).

# Extending the Default Worker without NPM

Sometimes it is useful to encapsulate commonly used Brigade code into a library
that can be shared between projects internally. While the NPM model above is
easier to manage over the longer term, there is a simple method for loading
custom code into an image. This section illustrates that method.

Here is a small library that adds an `alpineJob()` helper function:

[mylib.js](examples/workers/mylib.js)
```javascript
const {Job} = require("./brigadier");

exports.alpineJob = function(name) {
  j = new Job(name, "alpine:3.7", ["echo hello"])
  return j
}
```

We can build this file into our `Dockerfile` by copying it into the image:

```
FROM brigadecore/brigade-worker:v1.2.1

RUN yarn add xml-simple
COPY mylib.js /home/src/dist
```

Next, skip to the section on [building and publishing a custom worker image](#building-and-publishing-a-custom-worker-image).

Use this in `brigade.js` like so:

```javascript
const { events } = require("brigadier");
const XML = require("xml-simple");
const { alpineJob } = require("./mylib");

events.on("exec", () => {
  XML.parse("<say><to>world</to></say>", (e, say) => {
    console.log(`Saying hello to ${say.to}`);
  })

  const alpine = alpineJob("myjob");
  alpine.run();
});
```

# Creating a Completely Custom Worker (Advanced)

If you wish to create a custom worker that doesn't merely make new Node.js
libraries available to your `brigade.js` file, nothing prevents you from
defining a new class of worker from scratch. This approach allows significantly
more flexibility with respect to how the worker is implemented (languages,
frameworks, etc.) _and_ how the worker functions. For instance, it is entirely
possible to create workers that bypass `brigade.js` and drive builds based on
some other declarative or imperative format.

The remainder of this section covers the general requirements for a custom
worker as well as the methods whereby the Brigade controller passes
configuration to a worker.

## Environment Variables

The Brigade controller sets the following environment variables in the
Kubernetes pod that executes a worker. Use these environment variables to learn
about project-specific, and build/event-specific details.

Note that custom workers may selectively disregard any environment variables
they deem inapplicable to the custom behavior they implement. For instance, an
environment variable that conveys the expected location of the `brigade.js` file
is not applicable to a worker that does not utilize such configuration.

### Brigade Level Environment Variables

| Environment Variable Name | Description | Notes |
|---------------------------|-------------|-------|
| `BRIGADE_DEFAULT_BUILD_STORAGE_CLASS` | The Kubernetes StorageClass to use for shared build storage if shared build storage is required _and_ if no StorageClass is specified in project-level configuration. | Ignore this if your custom worker never uses shared build storage. |
| `BRIGADE_DEFAULT_CACHE_STORAGE_CLASS` | The Kubernetes StorageClass to use for caching jobs if build cache storage is required _and_ no StorageClass is specified in project-level configuration. | Ignore this if your custom worker never uses a build cache. |
| `BRIGADE_WORKSPACE` | If applicable, the location where project source code obtained from a VCS repository should be placed. The Brigade controller hardcodes this as `/vcs`. | Ignore this if your custom worker would like to place project source code elsewhere. |

### Project Level Environment Variables

| Environment Variable Name | Description | Notes |
|---------------------------|-------------|-------|
| `BRIGADE_CONFIG` | If applicable, may override the default location of the `brigade.json` configuration file. | |
| `BRIGADE_LOG_LEVEL` | Desired log level. | This is typically left unset by the controller. |
| `BRIGADE_PROJECT_ID` | A unique identifier for the Brigade project. | |
| `BRIGADE_PROJECT_NAMESPACE` | The Kubernetes namespace in which the worker should create any pods that implement each build's job(s). The  worker must have write access to this namespace. | Note this is always the same namespace as the one that the worker itself is executed. |
| `BRIGADE_REMOTE_URL` | If applicable, a URL for obtaining project source code from a VCS repository. | |
| `BRIGADE_REPO_AUTH_TOKEN` | If applicable, an authentication token for accessing the project's private source code repository. | |
| `BRIGADE_REPO_KEY` | If applicable, an ssh key for accessing the project's private source code repository. | |
| `BRIGADE_REPO_SSH_CERT` | If applicable, an ssh certificate used together with ssh key. | |
| `BRIGADE_SCRIPT` | If applicable, may override the default location of the `brigade.js` file. | |
| `BRIGADE_SECRET_KEY_REF` | A boolean (represented as the _string_ `"true"` or `"false"`) indicating whether pods that implement each build's job(s) may utilize `secretKeyRef` in defining their own environment variables. | |
| `BRIGADE_SERVICE_ACCOUNT` | The service account to be used by any pods that implement each build's job(s). | Note that this may be different from the service account used by the worker itself. | |
| `BRIGADE_SERVICE_ACCOUNT_REGEX` | If applicable, constrains which service accounts may be used any pods that implement each build's job(s). | |

### Build Level Environment Variables

| Environment Variable Name | Description | Notes |
|---------------------------|-------------|-------|
| `BRIGADE_BUILD_ID` | A unique identifier for the build. | |
| `BRIGADE_BUILD_NAME` | A unique name for the _worker_ handling the build. | |
| `BRIGADE_COMMIT_ID` | If applicable, the VCS commit ID. | For example, a git SHA. |
| `BRIGADE_COMMIT_REF` | If applicable, the a VCS reference. | For example, `refs/heads/master`. | |
| `BRIGADE_EVENT_PROVIDER` | The name of the gateway that was the source of the triggering event. | For example, `github` or `dockerhub`. |
| `BRIGADE_EVENT_TYPE` | The type of event that triggered the build. | For example, `push` or `pull_request`. | |

## Additional Project Level Configuration

Custom workers may obtain additional project-level configuration (not provided
by the controller as environment variables) by using the Kubernetes API to
retrieve the applicable project secret. Brigade stores this secret in the
namespace specified by the `BRIGADE_PROJECT_NAMESPACE` environment variable,
with a name specified by the `BRIGADE_PROJECT_ID`.

The following table summarizes _useful_ project-level configuration available
via this method. Configuration not applicable to custom workers or redundant
with the documented environment variables is omitted.

| Secret Key | Description | Notes |
|------------|-------------|-------|
| `allowHostMounts` | A boolean (represented as the _string_ `"true"` or `"false"`) indicating whether pods that implement each build's job(s) may mount paths from the underlying host. | |
| `allowPrivilegedJobs` | A boolean (represented as the _string_ `"true"` or `"false"`) indicating whether pods that implement each build's job(s) may include privileged containers. | |
| `buildStorageSize` | The desired size for any shared build storage and build cache volumes that are provisioned. | |
| `initGitSubmodules` | If applicable, a boolean (represented as the _string_ `"true"` or `"false"`) indicating whether any git submodules should be initialized after project source is retrieved from VCS. | |
| `kubernetes.buildStorageClass` | Specifies the desired Kubernetes storage class to be used for any shared build storage volume that is provisioned. | This can override the Brigade-level default. |
| `kubernetes.cacheStorageClass` | Specifies the desired Kubernetes storage class to be used for any build cache volume that is provisioned. | This can override the Brigade-level default. |
| `secrets` | Base64-encoded JSON containing project-specific secrets. | |
| `vcsSidecar` | If applicable, image to be used by "VCS sidecar" containers that obtain project source code from a VCS repository. | |

## Job Pod Names and Labels

To be visible to the rest of Brigade (and related projects, such as Kashti) and
recognizable as pods that implement a build's job(s), certain naming and
labeling conventions must be adhered to when a worker spawns such pods.

1. Pod names MUST take the form `<job name>-<build ID>`.

1. Pods MUST be labeled as follows:

    | Key | Value |
    |-----|-------|
    | `heritage` | `brigade` |
    | `component` | `job` |
    | `jobname` | Job name |
    | `project` | Project ID |
    | `worker` | Worker ID (aka build name) |
    | `build` | Build ID |

All the [usual constraints](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#syntax-and-character-set) in labeling Kubernetes pods apply.

## Exit Code

Brigade determines the ultimate success or failure of a build by the return code
from the worker.

Workers that successfully execute to completion with no errors MUST exit with
return code 0.

Worker executions that fail MUST exit with a non-zero return code.

# Building and Publishing a Custom Worker Image

Whether you are extending the default worker image or creating a worker entirely
from scratch. The next step is to build a Docker image from your `Dockerfile`:

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

As of Brigade v0.10.0, worker images can be configured _globally_. Individual
projects can choose to override the global setting.

To set the version globally, you should override the following values in your
`brigade/brigade` chart:

```yaml
# worker is the JavaScript worker. These are created on demand by the controller.
worker:
  registry: myregistry
  name: myworker
  tag: latest
  #pullPolicy: IfNotPresent # Set this to Always if you are testing and using
  #                           an upstream registry like Dockerhub or ACR
```

You can then use `helm upgrade` to load those new values to Brigade.

## Project Overrides

To configure the worker image per-project, you can set up a custom `worker` section
via `brig` during the `Configure advanced options` section.  (If the project has
already been created, use `brig project create --replace -p <pre-existing-project>`).

Here we supply our custom worker image registry (`myregistry`), image name
(`myworker`), image tag (`latest`), pull policy (`IfNotPresent`) and command (`yarn -s start`):

```console
$ brig project create
...
? Configure advanced options Yes
...
? Worker image registry or DockerHub org myregistry
? Worker image name myworker
? Custom worker image tag latest
? Worker image pull policy IfNotPresent
? Worker command yarn -s start
```

## Using Your Custom Worker

Once you have set the Docker image (above), your new Brigade workers will
automatically switch to using this new image.

# Best practices

We strongly discourage attempting to turn a worker into a long-running server.
This violates the design assumptions of Brigade, and can result in unintended
side effects.
