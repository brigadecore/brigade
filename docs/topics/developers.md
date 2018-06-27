# Developer Guide

This document explains how to get started developing Brigade.

Brigade is composed of numerous parts:

- brigade-controller: The Kubernetes controller for delegating Brigade events
- brigade-server: The implementation of the GitHub web hooks. It requires
  the controller.
- brigade-worker: The JavaScript runtime for executing `brigade.js` files. The
  controller spawns these, though you can run one directly as well.
- brigade-api: The REST API server for user interfaces
- brigade-project: The Helm chart for installing Brigade projects
- git-sidecar: The code that runs as a sidecar in cluster to fetch Git repositories

This document covers development of `brigade-controller`, `brigade-server`, and
`brigade-worker`.

## Prerequisites

- Go toolchain (latest version)
- Minikube
- Docker
- make
- Node.js and NPM

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

- `$ git checkout v0.15.0`

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
for Brigade to operate.

For local development, you will want to point your Docker client to the Minikube
Docker daemon:

```
$ eval $(minikube docker-env)
```

Running `VERSION=latest make docker-build` will push the Brigade images to the Minikube Docker
daemon. You can verify this by running `docker images`. You should see the `latest` tags for
the brigade images.

Now create a custom `values.yaml` file for the chart, and set the images to all
pull the `latest` image:

```
$ helm inspect values ./charts/brigade > myvalues.yaml
$ open myvalues.yaml    # Change all `tag:` fields to be `tag: latest`
```

From here, you can install Brigade into Minikube using the Helm chart:

```
$ helm install -n brigade ./charts/brigade -f myvalues.yaml # if this command fails run `helm init`
```

Don't forget to also create a project (`$ helm install -n empty-testbed charts/brigade-project`).

## Running Brigade inside remote Kubernetes

Some developers use a remote Kubernetes instead of minikube.

To run a development version of Brigade inside of a remote Kubernetes,
you will need to do two things:

- Make sure you push your `brigade` docker images to a registry the cluster can access
- Set the image when you do a `helm install ./chart` on the Brigade chart.

## Running Brigade (brigade-server) Locally (against Minikube)

Assuing you have Brigade installed (either on minikube or another cluster) and
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
flexible way of doing this is via the `./charts/brigade-project` Helm chart:

```console
$ helm inspect ./charts/brigade-project > functional-test-project.yaml
$ # edit the functional-test-project.yaml file
$ helm install -f functional-test-project.yaml -n brigade-functional-tests ./charts/brigade-project
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

Once Helm installs the project, you can test it with `helm get brigade-functional-tests`.

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
