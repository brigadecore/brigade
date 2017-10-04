# Installation Guide

This guide provides detailed information about installing and configuring the Acid services.

## Installing the Kubernetes Services

Acid runs inside of a Kubernetes cluster. This section explains how to install
into an existing Kubernetes cluster.

### Option 1: Install from GitHub Repository

To install from the GitHub repository, follow these steps:

```console
$ git clone https://github.com/deis/acid.git
$ cd acid
$ helm install ./chart/acid
```

### Option 2: Install from the Chart Repository

It is also possible to build from the Chart Repository. The repository is only
updated when a new version is released, and will likely not be as up-to-date as
the GitHub repository.

```console
$ helm repo add acid https://deis.github.io/acid
$ helm install acid/acid
```

### Customizing Installations

Both of these options use the Acid binaries stored in DockerHub. But you can override
this behavior by supplying an alternative image during install:

```console
$ helm install acid/acid --set image.name=my-image --set image.tag=1.2.3
```

There are a variety of other configuration options for Acid. Run `helm fetch values ./chart/acid`
to see them all.

### Disabling RBAC

By default, Acid has Role Based Access Control support. To disable this, set
`rbac.enabled` to `false`:

```console
$ helm install acid/acid --set rbac.enabled=false
```

## Configuring Acid

Once Acid is installed, three deployments will be running:

- deployments/acid-core-acid
- deployments/acid-core-acid-api
- deployments/acid-core-acid-ctrl

The controller (acid-core-acid-ctrl) is the primary piece of Acid. It listens for
new Acid events and triggers new builds.

### Configuring Persistent Volumes

Acid creates Persistent Volume Claims on the fly. By default it will create
PVCs using the default persistent volume class. Changing your cluster's default
storage class will change what Acid creates.

We recommend the following file system types for these distributions or platforms:

- Azure: AzureFile
- Minikube: 9p (the default)
- General: A storage type that supports ReadWriteMany.
