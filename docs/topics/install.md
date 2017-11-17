# Installation Guide

This guide provides detailed information about installing and configuring the Brigade services.

> Brigade is under highly active development. Day to day, the `master` branch is
> changing. If you choose to build from source, you may prefer to build off of
> a tagged release rather than master.

## Installing the Kubernetes Services

Brigade runs inside of a Kubernetes cluster. This section explains how to install
into an existing Kubernetes cluster.

### Option 1: Install from the Chart Repository

Each time the Brigade team cuts a new release, we update the Helm charts. Installing
with Helm is the best way to get a working release of Brigade.

```console
$ helm repo add brigade https://azure.github.io/brigade
$ helm install brigade/brigade
```

### Option 2: Install from GitHub Repository

If you are developing Brigade, or are interested in testing the latest features
(at the cost of additional time and energy), you can build Brigade from source.

The `master` branch typically contains newer code than the last release. However,
the charts will install the last released version.

```console
$ git clone https://github.com/azure/brigade.git
$ cd brigade
$ # optionally check out a tagged release: git checkout v0.5.0
$ helm install ./charts/brigade
```

Once you have Brigade installed, you can proceed to [creating a project](projects.md).
The remainder of this guide covers special configurations of Brigade.

> If you are not working off of a tagged release, you may also have to build
> custom images. The [Developers Guide](developers.md) explains this in more
> detail. Otherwise, the images referenced by the chart will be from the last
> release, and may not have the latest changes.

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
