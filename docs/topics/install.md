# Installation Guide

This guide provides detailed information about installing and configuring the Brigade services.

## Installing the Kubernetes Services

Brigade runs inside of a Kubernetes cluster. This section explains how to install
into an existing Kubernetes cluster.

### Option 1: Install from GitHub Repository

To install from the GitHub repository, follow these steps:

```console
$ git clone https://github.com/azure/brigade.git
$ cd brigade
$ helm install ./charts/brigade
```

### Option 2: Install from the Chart Repository

It is also possible to build from the Chart Repository. The repository is only
updated when a new version is released, and will likely not be as up-to-date as
the GitHub repository.

```console
$ helm repo add brigade https://azure.github.io/brigade
$ helm install brigade/brigade
```

Once you have Brigade installed, you can proceed to [creating a project](projects.md).
The remainder of this guide covers special configurations of Brigade.

### Customizing Installations

Both of these options use the Brigade binaries stored in DockerHub. But you can override
this behavior by supplying an alternative image during install:

```console
$ helm install brigade/brigade --set image.name=my-image --set image.tag=1.2.3
```

There are a variety of other configuration options for Brigade. Run `helm fetch values ./charts/brigade`
to see them all.

### Enabling RBAC (optional)

By default, Brigade has Role Based Access Control support turned off. To enable this, set
`rbac.enabled` to `true`:

```console
$ helm install brigade/brigade --set rbac.enabled=true
```

> RBAC is disabled by default because many clusters to not enable RBAC by default.

## Configuring Brigade

Once Brigade is installed, three deployments will be running:

- deployments/brigade-core-brigade
- deployments/brigade-core-brigade-api
- deployments/brigade-core-brigade-ctrl

The controller (brigade-core-brigade-ctrl) is the primary piece of Brigade. It listens for
new Brigade events and triggers new builds.

### Configuring Persistent Volumes

Brigade creates Persistent Volume Claims on the fly. By default it will create
PVCs using the default persistent volume class. Changing your cluster's default
storage class will change what Brigade creates.

We recommend the following file system types for these distributions or platforms:

- Azure: AzureFile
- Minikube: 9p (the default)
- General: A storage type that supports ReadWriteMany.
