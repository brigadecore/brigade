---
title: Developer Guide
description: How to get started developing Brigade
aliases:
  - /developers.md
  - /topics/developers.md
---

TODO: update per v2

# Developer Guide

This document explains how to get started developing Brigade.

Brigade is composed of numerous parts.  The following represent the core components:

- brigade-controller: The Kubernetes controller for delegating Brigade events
- brigade-worker: The JavaScript runtime for executing `brigade.js` files. The
  controller spawns these, though you can run one directly as well.
- brigade-api: The REST API server for user interfaces
- brigade-project: The Helm [chart][brigade-project-chart] for installing Brigade projects
- brigade-vacuum: The stale build cleaner-upper (optional; enabled by default)
- brig: The Brigade CLI
- git-sidecar: The code that runs as a sidecar in cluster to fetch Git repositories (optional; enabled by default)

Additionally, there are several opt-in gateways that can be enabled via Helm chart values.  These are:

 - Brigade GitHub App Gateway: The implementation of the GitHub App web hooks. It requires
  the controller.
 - Generic Gateway: A generic gateway offering flexibility to create Brigade events from webhooks
  originating from an arbitrary service/platform.
- Container Registry Gateway: A gateway supporting container registry webhooks such as the ones emitted by
DockerHub and ACR.

Read up on all the gateways above, as well as others, in the [Gateways doc](./gateways.md).

This document covers environment setup, how to run functional tests and development of
`brigade-controller` and `brigade-worker` components.

## Prerequisites

- Minikube or [kind](https://github.com/kubernetes-sigs/kind) (running k8s versions 1.16.0+)
- Docker
- make

## Clone the Repository In GOPATH

Building from source does not _require_ code to be on your `GOPATH` since all
builds are containerized by default, however, if you do have Go installed
locally and wish (for instance) for your text editor or IDE's Go support to work
properly with this project, then follow these optional steps for cloning the
Brigade repository into your `GOPATH`:

```console
$ export GOPATH=$(go env GOPATH) # GOPATH is set to $HOME/go by default
$ export PATH=$GOPATH/bin:$PATH
$ mkdir -p $GOPATH/src/github.com/brigadecore
$ git clone https://github.com/brigadecore/brigade $GOPATH/src/github.com/brigadecore/brigade
$ cd $GOPATH/src/github.com/brigadecore/brigade
```

**Note**: this leaves you at the tip of **master** in the repository where active development
is happening. You might prefer to checkout the most recent stable tag:

- `$ git checkout v1.2.1`

After cloning the project locally, you should run this command to [configure the remote](https://help.github.com/articles/configuring-a-remote-for-a-fork/): 

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
same results in CI.

## Developing on Windows

All development-related tasks should "just work" on Linux and Mac OS systems.
When developing on Windows, the maintainers strongly recommend utilizing the
Windows Subsystem for Linux.

[This blog post](https://nickjanetakis.com/blog/setting-up-docker-for-windows-and-wsl-to-work-flawlessly)
provides excellent guidance on making the Windows Subsystem for Linux work
seamlessly with Docker Desktop (Docker for Windows).

## Working with Go Code

To run lint checks:

```console
$ make lint
```

To format the Go files:

```console
$ make format-go
```

To run the unit tests:

```console
$ make test-unit
```

To re-run Go dependency resolution:

```console
$ make dep
```

## Working with JS Code (for the Brigade Worker)

To format the Javascript files:

```console
$ make format-js
```

To run the tests:

```console
$ make test-js
```

To re-run JS dependency resolution:

```console
$ make yarn-install
```

(See `Running the Brigade-Worker Locally` below for live testing against a running instance.)

## Building Source

To build all of the source, run this:

```console
$ make build-all-images build-brig
```

To build just the Docker images, run:

```console
$ make build-all-images
```

To build just the client binary for your OS, run this:

```console
$ make build-brig
```

To build all the supported client binaries (for Mac, Linux, and Windows on amd64), run
this:

```console
$ make xbuild-brig
```

## Pushing Images

By default, built images are named using the following scheme:
`<component>:<version>`. If you wish to push customized or experimental images
you have built from source to a particular org on a particular Docker registry,
this can be controlled with environment variables.

The following, for instance, will build images that can be pushed to the
`krancour` org on Dockerhub (the registry that is implied when none is
specified).

```console
$ DOCKER_ORG=krancour make build-all-images
```

To build for the `krancour` org on a different registry, such as `quay.io`:

```console
$ DOCKER_REGISTRY=quay.io DOCKER_ORG=krancour make build-all-images
```

Images built with names that specify registries and orgs for which you have
write access can be pushed using `make push-all-images`. Note that the
`build-all-images` target is a dependency for the `push-all-images` target, so
the build _and_ push processes can be accomplished together like so:

Note also that you _must_ be logged into the registry in question _before_
attempting this.

```console
$ DOCKER_REGISTRY=quay.io DOCKER_ORG=krancour make push-all-images
```

## Minikube configuration

Start Minikube. Your addons should look like this:

```console
$  minikube addons list
- addon-manager: enabled
- dashboard: disabled
- default-storageclass: enabled
- heapster: disabled
- ingress: enabled
- kube-dns: enabled
- registry: disabled
- registry-creds: disabled
```

Feel free to enable other addons, but the ones above are expected to be present
for Brigade to operate.

For local development, you will want to point your Docker client to the Minikube
Docker daemon:

```console
$ eval $(minikube docker-env)
```

Running `make build-all-images` will build the Brigade images using the Minikube
Docker daemon. The image tag will be derived from the git sha.  You can verify
this by running `docker images` and affirming these tagged images are listed.

Brigade charts are hosted in the separate [brigadecore/charts][charts]
repo, so we'll need to add the corresponding Helm repo locally:

```console
$ helm repo add brigade https://brigadecore.github.io/charts
"brigade" has been added to your repositories
```

If you just want to roll with the default chart values and let the Makefile set
the image tags appropriately, simply run:

```console
$ make helm-install
```

This will issue the appropriate command to create a new Brigade chart release on this cluster.

Note: `helm init` may be needed to get tiller up and running on the cluster, if not already started.

Note also: If you were specific about `DOCKER_ORG` and/or `DOCKER_REGISTRY` when
building images, you should also be specific when running `make helm-install`.
For instance:

```console
$ DOCKER_ORG=krancour make build-all-images helm-install
```

During active development, the overall flow might then look like this:

```console
$ # make code changes, commit
$ make build-all-images helm-upgrade
$ # (repeat)
$ # push to fork and create pull request
```

For finer-grained control over installation, do not use `make helm-upgrade`.
Instead, you may opt to create a custom `values.yaml` file for the chart and set
various values in addition to the latest image tags:

```console
$ helm inspect values brigade/brigade > myvalues.yaml
$ open myvalues.yaml    # Change all `registry:` and `tag:` fields as appropriate
```

From here, you can install Brigade into Minikube using the Helm chart:

```console
$ helm install -n brigade brigade/brigade -f myvalues.yaml
```

Don't forget to also create a project.  Check out [projects](./projects.md) to see how it's done.

## Developing brigade with kind

You can also use [kind](https://github.com/kubernetes-sigs/kind) for your day to day Brigade development workflow. Kind has a great quickstart that can be found [here](https://kind.sigs.k8s.io/docs/user/quick-start/).

- Run `kind create cluster` to create the cluster 
- Run `export KUBECONFIG="$(kind get kubeconfig-path --name="kind")"` to set KUBECONFIG to the file that was created by kind
- Install helm on the `kind` cluster by running `helm init`. The latest [Helm 3](https://github.com/helm/helm) version is recommended. However, if you're running Helm 2, check [here](https://helm.sh/docs/using_helm/#role-based-access-control) for proper Helm/Tiller installation instructions or use
```
# Only if you are using Helm 2
kubectl --namespace kube-system create serviceaccount tiller
kubectl create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller 
kubectl --namespace kube-system patch deploy tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}' 
```
- Run `DOCKER_ORG=brigadecore make build-all-images load-all-images` to build all images locally for the kind cluster
- Run `make helm-install` to install/upgrade Brigade onto the kind cluster. This is the command you should re-run to test your changes during your Brigade development workflow. If this command does not work, you probably need to run `helm repo add brigade https://brigadecore.github.io/charts`

When you're done, feel free to `kind delete cluster` to tear down the kind cluster resources.

## Running Brigade inside remote Kubernetes

Some developers use a remote Kubernetes instead of minikube.

To run a development version of Brigade inside of a remote Kubernetes,
you will need to do two things:

- Make sure you push your `brigade` docker images to a registry the cluster can access
- Set the image when you do a `helm install brigade/<chart>` on the Brigade chart.

## Running Brigade (brigade-controller) Locally (against Minikube or kind)

Assuming you have Brigade installed (either on minikube or another cluster) and
your `$KUBECONFIG` is pointing to that cluster, you can run `brigade-controller`
locally.

```console
$ ./bin/brigade-controller --kubeconfig $KUBECONFIG
```

(The default location for `$KUBECONFIG` on UNIX-like systems is `$HOME/.kube`.)

For the remainder of this document, we will assume that your local `$KUBECONFIG`
is pointing to the correct cluster.

### Running the Functional Tests

Once you have Brigade running in Minikube or a comparable alternative, you should be
able to run the functional tests.

First, create a project that points to the `brigadecore/empty-testbed` project. The most
flexible way of doing this is via the `brig` cli.  Here we supply `-x` to forgo
interactive prompts.  All the defaults will therefore be set to the
[brigadecore/empty-testbed](https://github.com/brigadecore/empty-testbed) project.

```console
 $ brig project create -x
Project ID: brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
```

You can check this project configuration out via `brig project get brigadecore/empty-testbed`.

With this setup, you should be able to run `make test-functional` and see the
tests run against your local Brigade binary.

## Running the Brigade-Worker Locally

You can run the Brigade worker locally by `cd`ing into `brigade-worker` and running
`k brigade`. Note that this will require you to set a number of environment
variables. See `brigade-worker/index.ts` for the list of variables you will need
to set.

Here is an example script for running a quick test against a locally running brigade worker.

```bash
#!/bin/bash

export BRIGADE_EVENT_TYPE=quicktest
export BRIGADE_EVENT_PROVIDER=script
export BRIGADE_COMMIT_REF=master
export BRIGADE_PAYLOAD='{}'
export BRIGADE_PROJECT_ID=brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
export BRIGADE_PROJECT_NAMESPACE=default
export BRIGADE_SCRIPT="$(pwd)/brigade.js"
export BRIGADE_CONFIG="$(pwd)/brigade.json"

cd ./brigade-worker
echo "running $BRIGADE_EVENT_TYPE on $BRIGADE_SCRIPT with $BRIGADE_CONFIG config for $BRIGADE_PROJECT_ID"
yarn start
```

You may change the variables above to point to the desired project.

[charts]: https://github.com/brigadecore/charts
[brigade-project-chart]: https://github.com/brigadecore/charts/tree/master/charts/brigade-project

> Note: an Node dependency audit is part of the build process. To execute it manually, before pushing, you can run `make yarn-audit`.

## End to end testing

We've written an end to end test scenario for Brigade that that you can run using `make e2e`. Currently, what the test in the `run.sh` does is

* installs kubectl, kind, helm 3 if not already installed
* builds docker images of Brigade components
* loads them into kind
* installs them onto the kind cluster
* confirm that all components are successfully deployed
* installs a test Brigade project (brig project create -x -f) and confirms that the corresponding k8s Secret is created
* runs a custom brigade.js and verifies some output from worker Pod
* on completion (or on error) it tears down the kind cluster