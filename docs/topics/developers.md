# Developer Guide

This document explains how to get started developing Acid.

Acid is composed of numerous parts:

- acid-controller: The Kubernetes controller for delegating Acid events
- acid-server: The implementation of the GitHub web hooks. It requires
  the controller.
- acid-worker: The JavaScript runtime for executing `acid.js` files. The
  controller spawns these, though you can run one directly as well.
- acid-api: The REST API server for user interfaces
- acid-project: The Helm chart for installing Acid projects
- vcs-sidecar: The code that runs as a sidecar in cluster to fetch VCS repositories

This document covers development of `acid-controller`, `acid-server`, and
`acid-worker`.

## Prerequisites

- Go toolchain (latest version)
- Minikube
- Docker
- make
- Node.js and NPM


## Building Source

To build all of the source, run this:

```
$ make bootstrap build
```

To build Docker images, run:

```
$ make docker-build
```

## Minikube configuration

Start Minikube. Your addons should look like this:

```
$  minikube addons list
- default-storageclass: enabled
- kube-dns: enabled
- dashboard: disabled
- heapster: disabled
- ingress: enabled
- registry: disabled
- registry-creds: disabled
- addon-manager: enabled
```

Feel free to enable other addons, but the ones above are expected to be present
for Acid to operate.

For local development, you will want to point your Docker client to the Minikube
Docker daemon:

```
$ eval $(minikube docker-env)
```

Running `make docker-build docker-push` will push the Acid images to the Minikube Docker
daemon.

## Running Acid inside remote Kubernetes

Some developers use a remote Kubernetes instead of minikube.

To run a development version of Acid inside of a remote Kubernetes,
you will need to do two things:

- Make sure you push your `acid` docker images to a registry the cluster can access
- Set the image when you do a `helm install ./chart` on the Acid chart.

## Running Acid (acid-server) Locally (against Minikube)

Assuing you have Acid installed (either on minikube or another cluster) and
your `$KUBECONFIG` is pointing to that cluster, you can run `acid` (acid-server)
locally.

```
$ ./bin/acid --kubeconfig $KUBECONFIG
```

(The default location for `$KUBECONFIG` on UNIX-like systems is `$HOME/.kube`.)

For the remainder of this document, we will assume that your local `$KUBECONFIG`
is pointing to the correct cluster.

### Running the Functional Tests

Once you have Acid running in Minikube or a comparable alternative, you should be
able to run the functional tests.

First, create a project that points to the `deis/empty-testbed` project. The most
flexible way of doing this is via the `./acid-project` Helm chart:

```console
$ helm inspect ./acid-project > functional-test-project.yaml
$ # edit the functional-test-project.yaml file
$ helm install -f functional-test-project.yaml -n acid-functional-tests ./acid-project
```

At the very least, you will want a config that looks like this:

```yamlproject: "deis/empty-testbed"
project: deis/empty-testbed
repository: "github.com/deis/empty-testbed"
cloneURL: "https://github.com/deis/empty-testbed.git"
namespace: "default"
```
It is possible to run the functional tests against a clone of the repo above,
but there's no need to. Basically we are testing GitHub connectivity and transactions
in these tests.

Once Helm installs the project, you can test it with `helm get acid-functional-tests`.

With this setup, you should be able to run `make test-functional` and see the
tests run against your local Acid binary.

## Running the Acid-Worker Locally

You can run the Acid worker locally by `cd`ing into `acid-worker` and running
`k acid`. Note that this will require you to set a number of environment
variables. See `acid-worker/index.ts` for the list of variables you will need
to set.

Here is an example script for running a quick test against a locally running acid worker.

```
#!/bin/bash

export ACID_EVENT_TYPE=quicktest
export ACID_EVENT_PROVIDER=script
export ACID_COMMIT=9c75584920f1297008118915024927cc099d5dcc
export ACID_PAYLOAD='{}'
export ACID_PROJECT_ID=acid-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
export ACID_PROJECT_NAMESPACE=default
export ACID_SCRIPT="$(pwd)/acid.js"

cd ./acid-worker
echo "running $ACID_EVENT_TYPE on $ACID_SCRIPT for $ACID_PROJECT_ID"
yarn start
```

You may change the variables above to point to the desired project.

