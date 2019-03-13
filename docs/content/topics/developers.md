---
title: Developer Guide
description: How to get started developing Brigad
---

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

- Go toolchain (latest version)
- Minikube
- Docker
- make
- Node.js, Yarn and NPM

## Clone the Repository In GOPATH

Follow these steps when cloning the brigade repository to use an existing `GOPATH` for your system:

```bash
export GOPATH=$(go env GOPATH) # GOPATH is set to $HOME/go by default
export PATH=$GOPATH/bin:$PATH # 'make bootstrap brig' will try to execute binnaries in $GOPATH/bin
mkdir -p $GOPATH/src/github.com/Azure
git clone https://github.com/Azure/brigade $GOPATH/src/github.com/Azure/brigade
cd $GOPATH/src/github.com/Azure/brigade
```

**Note**: this leaves you at the tip of **master** in the repository where active development
is happening. You might prefer to checkout the most recent stable tag:

- `$ git checkout v0.20.0`

After cloning the project locally, you should run this command to [configure the remote](https://help.github.com/articles/configuring-a-remote-for-a-fork/): 

```bash
git remote add fork https://github.com/<your GitHub username>/brigade
```

To push your changes to your fork, run:

```bash
git push --set-upstream fork <branch>
```

## Building Source

To build all of the source, run this:

```
$ make bootstrap build
```

To build just the client binaries, run this:

```
$ make bootstrap brig
```

To build Docker images, run:

```
$ make docker-build
```

## Javascript Bootstrap/Test

To bootstrap the Javascript dependencies required by Brigade Worker:

```
$ make bootstrap-js
```

To format the Javascript files:

```
$ make format-js
```

To run the tests:

```
$ make test-js
```

(See `Running the Brigade-Worker Locally` below for live testing against a running instance.)

## Minikube configuration

Start Minikube. Your addons should look like this:

```
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

```
$ eval $(minikube docker-env)
```

Running `make docker-build` will push the Brigade images to the Minikube Docker
daemon. The image tag (set by `VERSION` in the [Makefile](../../Makefile)) will default
to a unique value such as `v0.20.0-80-g6721dd8`.  You can verify this by running `docker images`
and affirming these tagged images are listed.

Brigade charts are hosted in the separate [Azure/brigade-charts][brigade-charts]
repo, so we'll need to add the corresponding Helm repo locally:

```
$ helm repo add brigade https://azure.github.io/brigade-charts
"brigade" has been added to your repositories
```

If you just want to roll with the default chart values and let the Makefile set
the image tags appropriately, simply run:

```console
$ make helm-install
```

This will issue the appropriate command to create a new Brigade chart release on this cluster.

(Note: `helm init` may be needed to get tiller up and running on the cluster, if not already started.)

During active development, the flow might then look like:

```console
$ # make code changes, commit
$ make docker-build
$ make helm-upgrade
$ # (repeat)
$ # push to fork and create pull request
```

For finer-grained control, you may wish to create a custom `values.yaml` file for the chart
and set various values in addition to the latest image tags:

```
$ helm inspect values brigade/brigade > myvalues.yaml
$ open myvalues.yaml    # Change all `tag:` fields to be `tag: <tag>`, etc.
```

From here, you can install Brigade into Minikube using the Helm chart:

```
$ helm install -n brigade brigade/brigade -f myvalues.yaml
```

Don't forget to also create a project.  Check out [projects](./projects.md) to see how it's done.

## Running Brigade inside remote Kubernetes

Some developers use a remote Kubernetes instead of minikube.

To run a development version of Brigade inside of a remote Kubernetes,
you will need to do two things:

- Make sure you push your `brigade` docker images to a registry the cluster can access
- Set the image when you do a `helm install brigade/<chart>` on the Brigade chart.

## Running Brigade (brigade-controller) Locally (against Minikube)

Assuming you have Brigade installed (either on minikube or another cluster) and
your `$KUBECONFIG` is pointing to that cluster, you can run `brigade-controller`
locally.

```
$ ./bin/brigade-controller --kubeconfig $KUBECONFIG
```

(The default location for `$KUBECONFIG` on UNIX-like systems is `$HOME/.kube`.)

For the remainder of this document, we will assume that your local `$KUBECONFIG`
is pointing to the correct cluster.

### Running the Functional Tests

Once you have Brigade running in Minikube or a comparable alternative, you should be
able to run the functional tests.

First, create a project that points to the `deis/empty-testbed` project. The most
flexible way of doing this is via the `brig` cli.  Here we supply `-x` to forgo
interactive prompts.  All the defaults will therefore be set to the
[deis/empty-testbed](https://github.com/deis/empty-testbed) project.

```console
 $ brig project create -x
Project ID: brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
```

You can check this project configuration out via `brig project get deis/empty-testbed`.

With this setup, you should be able to run `make test-functional` and see the
tests run against your local Brigade binary.

## Running the Brigade-Worker Locally

You can run the Brigade worker locally by `cd`ing into `brigade-worker` and running
`k brigade`. Note that this will require you to set a number of environment
variables. See `brigade-worker/index.ts` for the list of variables you will need
to set.

Here is an example script for running a quick test against a locally running brigade worker.

```
#!/bin/bash

export BRIGADE_EVENT_TYPE=quicktest
export BRIGADE_EVENT_PROVIDER=script
export BRIGADE_COMMIT_REF=master
export BRIGADE_PAYLOAD='{}'
export BRIGADE_PROJECT_ID=brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
export BRIGADE_PROJECT_NAMESPACE=default
export BRIGADE_SCRIPT="$(pwd)/brigade.js"

cd ./brigade-worker
echo "running $BRIGADE_EVENT_TYPE on $BRIGADE_SCRIPT for $BRIGADE_PROJECT_ID"
yarn start
```

You may change the variables above to point to the desired project.

[brigade-charts]: https://github.com/Azure/brigade-charts
[brigade-project-chart]: https://github.com/Azure/brigade-charts/tree/master/charts/brigade-project