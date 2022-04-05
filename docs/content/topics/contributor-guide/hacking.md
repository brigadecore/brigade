---
title: Hacking on Brigade
description: How to hack efficiently on Brigade
section: contributor-guide
weight: 3
---

This section is the primary technical primer on how to successfully make changes
to Brigade's code base, test those changes, and when necessary, build Brigade
from source and deploy it to a local development-grade Kubernetes cluster.

> ⚠️&nbsp;&nbsp;Most of this information is also generally applicable to other
> projects owned by the [@brigadecore](https://github.com/brigadecore) GitHub
> org.

> ⚠️&nbsp;&nbsp;__Special Note About Windows__
>
> All development-related tasks should "just work" on Linux and macOS systems.
> When developing on Windows, the maintainers strongly recommend utilizing the
> Windows Subsystem for Linux 2.  See more details
> [here](https://docs.docker.com/docker-for-windows/install/).

## Development Environment

Most Brigade components are implemented in Go. A few are implemented in
TypeScript. For maximum productivity in your text editor or IDE, it is
recommended that you have installed the latest stable releases of
[Go](https://go.dev/doc/install), [Node.js](https://nodejs.org/en/download/),
and applicable editor/IDE extensions, _however, this is not strictly required_
to be successful.

## Containerized Tests

In order to minimize the setup required to successfully apply small changes and
in order to reduce the incidence of "it worked on my machine," wherein changes
that pass tests locally do not pass the _same_ tests in CI due to environmental
differences, Brigade has adopted a "container-first" approach to testing. This
is to say we have made it the _default_ that unit tests, linters, and a variety
of other validations, when executed locally, automatically execute in a Docker
container that is maximally similar to the container in which those same tasks
will run during the continuous integration process.

To take advantage of this, you only need to have
[Docker](https://docs.docker.com/engine/install/) and `make` installed.

If you wish to opt-out of tasks automatically running inside a container, you
can set the environment variable `SKIP_DOCKER` to the value `true`. Doing so
will require that any tools involved in tasks you execute have been installed
locally.

## Working with Go Code

If you make modifications to Go code, it is recommended that you run
corresponding unit tests and linters before opening a PR.

To run lint checks for all Go-based components:

```shell
$ make lint-go
```

To run unit tests for all Go-based components:

```shell
$ make test-unit-go
```

## Working with TypeScript Code

If you make modifications to TypeScript code, it is recommended that you run
corresponding unit tests, style checks, and linters before opening a PR.

> ⚠️&nbsp;&nbsp;We use [Prettier](https://prettier.io/) to enforce consistent
> syntax/style and linters to catch potential problems that aren't directly
> syntax/style-related.

To run style checks for all TypeScript-based components:

```shell
$ make style-check-js
```

If this turns up any issues, you can correct them automatically by running:

```shell
$ make style-fix-js
```

To run lint checks for all TypeScript-based components:

```shell
$ make lint-js
```

To run unit tests for all TypeScript-based components:

```shell
$ make test-unit-js
```

## Building & Pushing Docker Images from Source

You will rarely, if ever, need to directly / manually build Docker images from
source. This is because of tooling we use (see next section) that does this for
you. Unless you have a specific need for doing this, you can safely skip this
section.

In the event that you do need to manually build images from source you _can_
execute the same make targets that are used by CI and our release process, but
be advised that this involves
[multiarch builds using buildx](https://www.docker.com/blog/multi-arch-build-and-images-the-simple-way/).
This can be somewhat slow and is not guaranteed to be supported on all systems.

First, list all available builders:

```shell
$ docker buildx ls
```

You will require a builder that lists both `linux/amd64` and `linux/arm64` as
supported platforms. If one is present, select it using the following command:

```shell
$ docker buildx use <NAME/NODE>
```

If you do not have an adequate builder available, you can try to launch one:

```shell
$ docker buildx create --use 
```

Because buildx utilizes a build server, the images built will not be present
locally. (Even though your build server is running locally, it's remote from the
perspective of your local Docker engine.) To make them available for use, you
_must_ push them somewhere. The following environment variables give you control
over _where_ the images are pushed to:

* `DOCKER_REGISTRY`: Host name of an OCI registry. If this is unset, Docker Hub
  is assumed.

* `DOCKER_ORG`: For multi-tenant registries, set this to a username or
  organization name for which you have permission to push images. This is not
  always required for private registries, but if you're pushing to Docker Hub,
  for instance, you will want to set this.

If applicable, you MUST log in to whichever registry you are pushing images to
in advance.

The example below shows how to build a single component and push it to Docker
Hub:

```shell
$ DOCKER_ORG=<Docker Hub username or org name> make push-<component name>
```

In this example, we push to a local registry instead:

```shell
$ DOCKER_REGISTRY=localhost:5000 make push-<component name>
```

To build and push _all_ components:

```shell
$ <env vars> make push-images
```

## Building the CLI

If you would like to build the `brig` CLI (command line interface) from
source using the same process that is used during CI and during release, you can
execute:

```shell
$ make build-cli
```

The commands above will build the CLI for a variety of OSes and
CPU architectures, which, cumulatively, can take quite some time.
If you would like to bypass this and build the CLI for you native OS and
operating system only, run the following instead:

```shell
$ make hack-build-cli
```

## Iterating Quickly

This section focuses on the best approaches for gaining rapid feedback on
changes you make to Brigade's code base.

By far, the fastest path to learning whether changes you have applied work as
desired is to execute unit tests as described in previous sections. If, however,
the changes you are applying are not well-covered by unit tests, it can become
advantageous to build Brigade from source, including your changes, and deploy it
to a live Kubernetes cluster. After doing so, you can execute our integration
test suite or you can test changes manually. Under these circumstances, a
pressing question is one of how Brigade can be built/re-built and deployed as
quickly as possible.

Building and deploying Brigade as quickly as possible requires minimizing the
the process' dependency on remote systems -- including Docker registries and
Kubernetes. To that end, we recommend a specific configuration wherein Docker
images are built and pushed to a _local_ image registry and a _local_ Kubernetes
cluster is configured such that it can pull images from that local registry.

Brigade's maintainers have never elected lightly to incorporate new third-party
tools into our recommended development processes, since we've learned from years
of experience that requiring any _extensive_ development environment setup can
be a source of frustration for would-be contributors. Even so, we have
identified three tools that, combined, have streamlined Brigade development to
such an extreme extent that they've become part of our recommended development
environment.

To continue, you will need to install the latest stable versions of:

* [KinD](https://kind.sigs.k8s.io/#installation-and-usage): Runs
  development-grade Kubernetes clusters in Docker.

* [ctlptl](https://github.com/tilt-dev/ctlptl#how-do-i-install-it): Launches
  development-grade Kubernetes clusters (in KinD, for instance) that are
  pre-connected to a local image registry.

* [Tilt](https://docs.tilt.dev/#macoslinux): Builds components from source and
  deploys them to a development-grade Kubernetes cluster. More importantly, it
  enables developers to rebuild and replace running components with the click of
  a button.

Follow the installation instructions for each of the above.

To launch a brand new Kind cluster pre-connected to a local image registry:

```shell
$ make hack-kind-up
```

To build and deploy all of Brigade from source:

```shell
$ tilt up
```

Tilt will also launch a web-based UI running at
[http://localhost:10350](http://localhost:10350). Visit this in your web browser
and you will be able to see the build and deployment status of each Brigade
component, complete with logs. Once Tilt has all of Brigade up and running, the
Brigade API server will be exposed (without TLS) on `localhost:31600`, so if you
wish to log in using the `brig` CLI as the "root" user:

```shell
$ brig login --server http://localhost:31600 --root
```

The root user's password is `F00Bar!!!`.

> ⚠️&nbsp;&nbsp;Tilt is often configured to watch files and automatically
> rebuild and replace running components when their source code is changed. _We
> have deliberately disabled this._ Each of our components takes long enough to
> build that we have discovered it's better for our CPUs if things aren't
> _constantly_ building and rebuilding in the background and build instead only
> when we choose. The web UI makes it easy to identify components whose source
> has been altered. They can be rebuilt and replaced with one mouse click.

When you are done with Tilt, interrupt the running `tilt up` process with
`ctrl + c`. Components _will remain running in the cluster_, but Tilt will no
longer be in control. If Tilt is restarted later, it will retake control of the
already-running components.

If you wish to undeploy everything Tilt has deployed for you, use `tilt down`.

To destroy your KinD cluster, use `make hack-kind-down`.

> ⚠️&nbsp;&nbsp;`make hack-kind-down` deliberately leaves your local registry
> running so that if you resume work later, you are doing so with a local
> registry that's already primed with most layers of each of Brigade's images.
> If you wish to destroy the registry, use:
>
> ```shell
> $ docker rm -f brigade-github-gateway-dev-cluster-control-plane
> ```
