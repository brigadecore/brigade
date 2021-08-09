---
title: Developer Guide
description: How to get started developing Brigade
section: topics
weight: 7
aliases:
  - /developers.md
  - /topics/developers.md
---

# Developer Guide

This document explains how to get started developing Brigade.

Brigade is composed of numerous parts.  The following represent the core
components:

- apiserver: Brigade's API server, where the majority of core Brigade
  logic lives
- scheduler: Schedules event workers and jobs on the underlying substrate
- observer: Observes (and records) state changes in workers and jobs and
  ultimately cleans them up from the underlying substrate
- worker: The default runtime for executing `brigade.js/.ts` files.
- logger: Event log aggregator, with Linux and Windows variants
- brig: The Brigade CLI
- git-initializer: The code that runs as a sidecar to fetch Git repositories
  for vcs-enabled projects

This document covers environment setup, how to run tests and development of
core components.

## Prerequisites

- A local Kubernetes cluster, 1.16.0+.  We recommend [kind] or [minikube].
- [Docker] _or_ [Go 1.16+] and JS dependencies
- make

[kind]: https://github.com/kubernetes-sigs/kind
[minikube]: https://github.com/kubernetes/minikube
[Docker]: https://www.docker.com/

## Clone the Repository

The first step to begin developing on Brigade is to clone the repo locally,
if you haven't already done so.  First navigate to your preferred work
directory and then issue the clone command:

```console
# Clone via ssh:
$ git clone git@github.com:brigadecore/brigade.git

# Or, via https:
$ git clone https://github.com/brigadecore/brigade.git
```

**Note**: this leaves you at the tip of **master** in the repository. As of
writing, active development on Brigade v2 is happening on the `v2` branch. To
switch branches, run the following command:

```console
$ git checkout v2
```

After cloning the project locally, you should run this command to
[configure the remote](https://help.github.com/articles/configuring-a-remote-for-a-fork/): 

```console
$ git remote add fork https://github.com/<your GitHub username>/brigade
```

To push your changes to your fork, run:

```console
$ git push --set-upstream fork <branch>
```

## Containerized Development Environment

To ensure a consistent development environment for all contributors, Brigade
relies heavily on Docker containers as sandboxes for all development activities
including dependency resolution, executing tests, or running a development
server.

`make` targets seamlessly handle the container orchestration.

If, for whatever reason, you must opt-out of executing development tasks within
containers, set the `SKIP_DOCKER` environment variable to `true`, but be aware
that by doing so, the success or failure of development-related tasks, tests,
etc. will be dependent on the state of your system, with no guarantee of the
same results in CI.  For running the full array of targets in this mode, you'll
need [Go 1.16+] and JavaScript ecosystem utilities ([yarn], [npm], [node],
etc.)

[Go 1.16+]: https://golang.org/doc/go1.16
[yarn]: https://yarnpkg.com/
[npm]: https://docs.npmjs.com/cli/v7/commands/npm
[node]: https://nodejs.org/en/

## Developing on Windows

All development-related tasks should "just work" on Linux and Mac OS systems.
When developing on Windows, the maintainers strongly recommend utilizing the
Windows Subsystem for Linux 2.  See more details
[here](https://docs.docker.com/docker-for-windows/install/).

## Working with Go Code

To run lint checks:

```console
$ make lint-go
```

To run the unit tests:

```console
$ make test-unit-go
```
## Working with JS Code (for the Brigade Worker)

To lint the Javascript files:

```console
$ make lint-js
```

To run the tests:

```console
$ make test-unit-js
```

To clear the JS dependency cache:

```console
$ make clean-js
```

## Building Source

Note that we use [kaniko]-based images to build Brigade component Docker
images. This is a convenient approach for running the same targets in CI
where configuring access to a Docker daemon is either unwieldy or rife with
security concerns. This also means that images built in this manner are not
preserved in the local Docker image cache. For the image push targets discussed
further down, the images are re-built with the same kaniko image, albeit with
the corresponding publishing flags added.

To build all of the source, run:

```console
$ make build
```

To build just the Docker images, run:

```console
$ make build-images
```

To build images via Docker directly and preserve images in the Docker cache,
run:

```console
$ make hack-build-images
```

To build all of the supported client binaries (for Mac, Linux, and Windows on
amd64), run:

```console
$ make build-cli
```

To build only the client binary for your current environment, run:

```console
$ make hack-build-cli
```

[kaniko]: https://github.com/GoogleContainerTools/kaniko

## Pushing Images

By default, built images are named using the following scheme:
`brigade-<component>:<version>`. If you wish to push customized or experimental
images you have built from source to a particular org on a particular Docker
registry, this can be controlled with environment variables.

The following, for instance, will build images that can be pushed to the
`krancour` org on Dockerhub (the registry that is implied when none is
specified). Here we use the targets that utilize Docker directly, so that the
images exist in the local cache: 

```console
$ DOCKER_ORG=krancour make hack-build-images
```

To build for the `krancour` org on a different registry, such as `quay.io`:

```console
$ DOCKER_REGISTRY=quay.io DOCKER_ORG=krancour make hack-build-images
```

Now you can push these images.  Note also that you _must_ be logged into the
registry in question _before_ attempting this.

```console
$ make hack-push-images
```

Otherwise, when using the kaniko-based targets, the images will be built and
pushed in one go.  Be sure to export the same registry/org values as above:

```console
$ DOCKER_REGISTRY=quay.io DOCKER_ORG=krancour make push-images
```

## Minikube configuration

Start Minikube with the following required addons enabled:

  - default-storageclass
  - storage-provisioner

To view all Minikube addons:

```console
$ minikube addons list
```

Additionally, for local development, it will be efficient to enable the
`registry` addon to set up a local registry to push images to.  See full
details in the [registry addon docs].  Here is an example on how to enable
the addon and redirect port 5000 on the Docker VM over to the Minikube machine:

```console
$ minikube addons enable registry
$ docker run -d --rm --network=host alpine ash \
  -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000"
```

Now to build and push images to the local registry and deploy Brigade, simply
run:

```console
$ export DOCKER_REGISTRY=localhost:5000
$ make hack
```

During active development, the overall flow might then look like this:

```console
$ # make code changes, commit
$ make hack
$ # (repeat)
$ # push to fork and create pull request
```

For finer-grained control over installation, you may opt to create a custom
`values.yaml` file for the chart and set various values in addition to the
latest image tags:

```console
$ helm inspect values charts/brigade > myvalues.yaml
$ open myvalues.yaml    # Change all `registry:` and `tag:` fields as appropriate
```

From here, you can install or upgrade Brigade into Minikube using the Helm
directly:

```console
$ helm upgrade --install -n brigade brigade charts/brigade -f myvalues.yaml
```

To expose the apiserver port, run the following command:

```console
$ make hack-expose-apiserver
```

You can then log in to the apiserver with the following `brig` command:

```console
$ brig login -s https://localhost:7000 -r -p 'F00Bar!!!' -k
```

To create your first Brigade project, check out [projects](./projects.md) to
see how it's done.

[registry addon docs]: https://minikube.sigs.k8s.io/docs/handbook/registry

## Kind configuration

You can also use [kind] for your day-to-day Brigade development workflow.
Kind has a great quickstart that can be found
[here](https://kind.sigs.k8s.io/docs/user/quick-start/).

The Brigade maintainers use kind heavily and there currently exists a helper
script that will create a new kind cluster with a local private registry
enabled. It also configures nfs as the local storage provisioner and maps
`localhost:31600` to the API server port. To use this script, run:

```console
$ make hack-new-kind-cluster
```

Now you're ready to build and push images to the local registry and deploy
Brigade:

```console
$ export DOCKER_REGISTRY=localhost:5000
$ make hack
```

You can then log in to the apiserver with the following `brig` command:

```console
$ brig login -s https://localhost:31600 -r -p 'F00Bar!!!' -k
```

To create your first Brigade project, check out [projects](./projects.md) to
see how it's done.

When you're done, if you'd like to clean up the kind cluster and registry
resources, run the following commands:

```console
$ kind delete cluster
$ docker rm -f kind-registry
```

## Running Brigade inside a remote Kubernetes cluster

Some developers use a remote Kubernetes cluster instead of Minikube or kind.

To run a development version of Brigade inside of a remote Kubernetes cluster,
you will need to do two things:

- Make sure you push the Brigade Docker images to a registry the cluster can
  access
- Export the correct values for `DOCKER_REGISTRY` and/or `DOCKER_ORG` prior
  to running `make hack`

## Running the Integration Tests

Once you have Brigade running in a Kubernetes cluster, you should be able to
run the integration tests, which will verifies basic Brigade functionality via
the following checks:

  - Logs into the apiserver
  - Creates projects of varying types
  - Creates an event for each project
  - Asserts worker and job logs and statuses for each event

To run the tests, issue the following command:

```console
$ make test-integration
```

See the [tests](./tests) directory to view all of the current integration
tests.