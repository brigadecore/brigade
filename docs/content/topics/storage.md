---
title: Storage
description: 'How Brigade uses Kubernetes Persistent Storage.'
---

# How Brigade uses Kubernetes Persistent Storage

Brigade allows script authors to declare two kinds of storage:

- per-job caches, which persist across builds
- per-build shared storage, which exists as long as the build is running

Usage of these is described within the [JavaScript docs](javascript.md) and the
[scripting guide](scripting.md).

This document describes the underlying Kubernetes architecture of these two
storage types.

## Brigade and PersistentVolumeClaims

Brigade provisions storage using Kubernetes PVCs. Both caches and shared storage
are PVC-backed.

### Caches

For a Cache, the Brigade worker will check to see if a Job asks for a cache. If it
does, the worker will create a PVC (if it doesn't already exist) and then mount
it to the cache.

> A Job, in this case, gains its identity from its name, and the project that
> it belongs to. So two hooks in the same brigade.js can redeclare a job name and
> thus share the cache.

That PVC is never removed by Brigade. Each subsequent run of the same Job will
then mount that same PVC.

### Shared Storage

Shared storage provisioning is markedly different than caches.

- The worker will _always_ provision a shared storage PVC _per build_.
- Each job _may_ mount this shared storage by setting its `storage.enabled` flag
  to `true`.
- At the end of a build, the storage will be destroyed.

In the current implementation, both the `after` and `error` hooks may attach to
the shared storage volume.

## Supporting Brigade Storage

Only certain volume plugins _can_ support Brigade. Specifically, **a volume driver
must be readWriteMany** in order for Brigade to use it. At the time of writing
very few VolumePlugins support the `readWriteMany` access mode. Ensure that your
volume plugin can support `readWriteMany`
([table](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes))
or that you're able to use [NFS](#using-an-nfs-server).

Only the following volume drivers are tested:

- Minikube's 9P implementation
- Azure's AzureFile storage
- NFS

We believe Gluster will work, but it's untested.

## Examples

### Using an NFS Server

As Brigade uses storage for caching and short-term file sharing, it is often convenient
to use storage backends that are optimized for short-term ephemeral storage.

NFS (Network File System) is one protocol that works well for Brigade. You can
use the [NFS Provisioner](https://github.com/IlyaSemenov/nfs-provisioner-chart)
chart to easily install an NFS server.

```console
$ helm repo add nfs-provisioner https://raw.githubusercontent.com/IlyaSemenov/nfs-provisioner-chart/master/repo
$ helm install --name nfs nfs-provisioner/nfs-provisioner --set hostPath=/var/run/nfs-provisioner
```

(Note that RBAC is enabled by default. To turn it off, use `--set rbac.enabled=false`.)

To use an emptyDir instead of a host mount, set `--hostPath=""`, like so:

```console
$ helm install --name nfs nfs-provisioner/nfs-provisioner --set hostPath=""
```

If you have plenty of memory to spare, and are more concerned with fast storage,
you can configure the provisioner to use a tmpfs in-memory filesystem.

```console
$ helm install --name nfs nfs-provisioner/nfs-provisioner --set hostPath="" --set useTmpfs=true
```

This chart installs a `StorageClass` named `local-nfs`. Brigade projects can
each declare which storage classes they want to use. And there are two storage
class settings:

- `kubernetes.cacheStorageClass`: This is used for the Job cache.
- `kubernetes.buildStorageClass`: This is used for the shared per-build storage.

In your project's `values.yaml` file, set both of those to `local-nfs`, and then
upgrade your project:

values.yaml
```yaml
# ...
kubernetes
kubernetes:
  buildStorageClass: local-nfs
  cacheStorageClass: local-nfs
```

Then:

```console
$ helm upgrade my-project brigade/brigade-broject -f values.yaml
```

If you would prefer to use the NFS provisioner as a cluster-wide default volume provider
(and have Brigade automatically use it), you can do so by making it the default
storage class:

```console
$ helm install --name nfs nfs-provisioner/nfs-provisioner --set hostPath="" --set defaultClass=true
```

Because Brigade pipelines can set up and tear down an NFS PVC very fast, the easiest
way to check that the above works is to run a `brig run` and then check the
log files for the NFS provisioner:

```console
$ kubectl logs nfs-provisioner-0 | grep volume
I0305 21:20:28.187133       1 controller.go:786] volume "pvc-06e2d938-20bb-11e8-a31a-080027a443a9" for claim "default/brigade-worker-01c7w0jse5grpkzwesz3htnnv5-master" created
I0305 21:20:28.195955       1 controller.go:803] volume "pvc-06e2d938-20bb-11e8-a31a-080027a443a9" for claim "default/brigade-worker-01c7w0jse5grpkzwesz3htnnv5-master" saved
I0305 21:20:28.195972       1 controller.go:839] volume "pvc-06e2d938-20bb-11e8-a31a-080027a443a9" provisioned for claim "default/brigade-worker-01c7w0jse5grpkzwesz3htnnv5-master"
I0305 21:20:34.208355       1 controller.go:1028] volume "pvc-06e2d938-20bb-11e8-a31a-080027a443a9" deleted
I0305 21:20:34.216852       1 controller.go:1039] volume "pvc-06e2d938-20bb-11e8-a31a-080027a443a9" deleted from database
I0305 21:21:15.967959       1 controller.go:786] volume "pvc-235dd152-20bb-11e8-a31a-080027a443a9" for claim "default/brigade-worker-01c7w0m8jw1h44vwhvzp4pr2dr-master" created
I0305 21:21:15.973328       1 controller.go:803] volume "pvc-235dd152-20bb-11e8-a31a-080027a443a9" for claim "default/brigade-worker-01c7w0m8jw1h44vwhvzp4pr2dr-master" saved
I0305 21:21:15.973358       1 controller.go:839] volume "pvc-235dd152-20bb-11e8-a31a-080027a443a9" provisioned for claim "default/brigade-worker-01c7w0m8jw1h44vwhvzp4pr2dr-master"
I0305 21:21:26.045133       1 controller.go:1028] volume "pvc-235dd152-20bb-11e8-a31a-080027a443a9" deleted
I0305 21:21:26.052593       1 controller.go:1039] volume "pvc-235dd152-20bb-11e8-a31a-080027a443a9" deleted from database
I0305 21:25:40.845601       1 controller.go:786] volume "pvc-c13e95f0-20bb-11e8-a31a-080027a443a9" for claim "default/brigade-worker-01c7w0wbffk3xhmbwwq114g15v-master" created
I0305 21:25:40.853759       1 controller.go:803] volume "pvc-c13e95f0-20bb-11e8-a31a-080027a443a9" for claim "default/brigade-worker-01c7w0wbffk3xhmbwwq114g15v-master" saved
I0305 21:25:40.853790       1 controller.go:839] volume "pvc-c13e95f0-20bb-11e8-a31a-080027a443a9" provisioned for claim "default/brigade-worker-01c7w0wbffk3xhmbwwq114g15v-master"
I0305 21:25:50.974719       1 controller.go:786] volume "pvc-c746f068-20bb-11e8-a31a-080027a443a9" for claim "default/github-com-deis-empty-testbed-three" created
I0305 21:25:50.994219       1 controller.go:803] volume "pvc-c746f068-20bb-11e8-a31a-080027a443a9" for claim "default/github-com-deis-empty-testbed-three" saved
I0305 21:25:50.994237       1 controller.go:839] volume "pvc-c746f068-20bb-11e8-a31a-080027a443a9" provisioned for claim "default/github-com-deis-empty-testbed-three"
I0305 21:25:56.974297       1 controller.go:1028] volume "pvc-c13e95f0-20bb-11e8-a31a-080027a443a9" deleted
I0305 21:25:56.985432       1 controller.go:1039] volume "pvc-c13e95f0-20bb-11e8-a31a-080027a443a9" deleted from database
```

Implementation details of note:

- The NFS server used is [NFS-Ganesha](https://github.com/nfs-ganesha/nfs-ganesha)
- The Kubernetes provisioner is part of [kubernetes-incubator/external-storage](https://github.com/kubernetes-incubator/external-storage/tree/master/nfs)
- Some Linux distros may not have the core NFS libraries installed. In such cases,
  NFS-Ganesha may not work. You may need to do something like `apt-get install nfs-common`
  on the nodes to install the appropriate libraris.

### Azure File Setup

If one has a Kubernetes cluster on [Azure](https://azure.microsoft.com/en-us/services/kubernetes-service/),
and the `default` storageclass is of the non-`readWriteMany`-compatible `kubernetes.io/azure-disk` variety, one can create
an Azure File storageclass and then configure the Brigade project to use this instead of `default`.

See the official [Azure File storageclass example](https://kubernetes.io/docs/concepts/storage/storage-classes/#azure-file)
for the yaml to use.  _(Hint: The parameters section can be omitted altogether and Azure will use the defaults associated
with the existing Kubernetes cluster.)_

Create the resource via `kubectl create -f azure-file-storage-class.yaml`.

Finally, be sure to set `kubernetes.buildStorageClass=azurefile` on the Brigade project Helm release, or via the "Advanced" set up
if creating via the `brig` cli.


## Errata

- At this point, cache PVCs are never destroyed, even if the project to which
  they belong is destroyed. This behavior may change in the future.
- Killing the worker pod will orphan shared storage PVCs, as the cleanup routine
  is part of the worker's shutdown process. If you manually destroy a worker pod,
  you must also manually destroy the associated PVCs.
