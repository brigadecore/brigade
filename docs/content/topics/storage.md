---
title: Storage
description: 'How Brigade uses Kubernetes Persistent Storage.'
aliases:
  - /storage.md
  - /intro/storage.md
  - /topics/storage.md
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
use the [NFS Server Provisioner](https://github.com/helm/charts/tree/master/stable/nfs-server-provisioner)
chart to easily install an NFS server.

```console
$ helm repo add stable https://kubernetes-charts.storage.googleapis.com/
$ helm install --name nfs stable/nfs-server-provisioner
```

By default, the chart installs with persistance disabled. For various methods on enabling, as well as
configuring other aspects of the installation, see the
[README](https://github.com/helm/charts/tree/master/stable/nfs-server-provisioner).

This chart installs a `StorageClass` named `nfs`. There are two options to configure Brigade
to use this storage class: at the server-level (for all Brigade projects) or at the project level.
Note that project-level settings will override the server-level settings.

In either case, there are two storage class settings:

- `cacheStorageClass`: This is used for the Job cache.
- `buildStorageClass`: This is used for the shared per-build storage.

To set these at the Brigade server level, set the values below accordingly in the Brigade
chart's `values.yaml` file:

values.yaml
```yaml
worker:
  defaultBuildStorageClass: nfs
  defaultCacheStorageClass: nfs
```

Then:

```console
$ helm upgrade brigade brigade/brigade -f values.yaml
```

To set these at the Brigade project level, set the values below accordingly in the Brigade
Project chart's `values.yaml` file:

values.yaml
```yaml
kubernetes:
  buildStorageClass: nfs
  cacheStorageClass: nfs
```

Then:

```console
$ helm upgrade my-project brigade/brigade-broject -f values.yaml
```

Note: The project-level settings can also be configured during the "Advanced" set up if creating via the `brig` CLI:

```console
 $ brig project create
...

? Build storage class nfs
? Job cache storage class  [Use arrows to move, type to filter, ? for more help]
  azurefile
  default
  managed-premium
❯ nfs
  Leave undefined
```

If you would prefer to use the NFS provisioner as a cluster-wide default volume provider
(and have Brigade automatically use it), you can do so by making it the default
storage class:

```console
$ helm install --name nfs stable/nfs-server-provisioner --set storageClass.defaultClass=true
```

Because Brigade pipelines can set up and tear down an NFS PVC very fast, the easiest
way to check that the above works is to run a `brig run` and then check the
log files for the NFS provisioner:

```console
$ kubectl logs -f nfs-nfs-server-provisioner-0
...

I1220 20:22:36.699672       1 controller.go:926] provision "default/brigade-worker-01dwjfkm36xf5539dh41fw9qsd" class "nfs": started
I1220 20:22:36.714277       1 event.go:221] Event(v1.ObjectReference{Kind:"PersistentVolumeClaim", Namespace:"default", Name:"brigade-worker-01dwjfkm36xf5539dh41fw9qsd", UID:"7535d094-2366-11ea-a145-72982b3e8f81", APIVersion:"v1", ResourceVersion:"8941502", FieldPath:""}): type: 'Normal' reason: 'Provisioning' External provisioner is provisioning volume for claim "default/brigade-worker-01dwjfkm36xf5539dh41fw9qsd"
I1220 20:22:36.727753       1 provision.go:439] using service SERVICE_NAME=nfs-nfs-server-provisioner cluster IP 10.0.241.101 as NFS server IP
I1220 20:22:36.739746       1 controller.go:1026] provision "default/brigade-worker-01dwjfkm36xf5539dh41fw9qsd" class "nfs": volume "pvc-7535d094-2366-11ea-a145-72982b3e8f81" provisioned
I1220 20:22:36.739785       1 controller.go:1040] provision "default/brigade-worker-01dwjfkm36xf5539dh41fw9qsd" class "nfs": trying to save persistentvolume "pvc-7535d094-2366-11ea-a145-72982b3e8f81"
I1220 20:22:36.749077       1 controller.go:1047] provision "default/brigade-worker-01dwjfkm36xf5539dh41fw9qsd" class "nfs": persistentvolume "pvc-7535d094-2366-11ea-a145-72982b3e8f81" saved
I1220 20:22:36.749113       1 controller.go:1088] provision "default/brigade-worker-01dwjfkm36xf5539dh41fw9qsd" class "nfs": succeeded
I1220 20:22:36.749196       1 event.go:221] Event(v1.ObjectReference{Kind:"PersistentVolumeClaim", Namespace:"default", Name:"brigade-worker-01dwjfkm36xf5539dh41fw9qsd", UID:"7535d094-2366-11ea-a145-72982b3e8f81", APIVersion:"v1", ResourceVersion:"8941502", FieldPath:""}): type: 'Normal' reason: 'ProvisioningSucceeded' Successfully provisioned volume pvc-7535d094-2366-11ea-a145-72982b3e8f81
I1220 20:22:43.083639       1 controller.go:1097] delete "pvc-7535d094-2366-11ea-a145-72982b3e8f81": started
I1220 20:22:43.089786       1 controller.go:1125] delete "pvc-7535d094-2366-11ea-a145-72982b3e8f81": volume deleted
I1220 20:22:43.116980       1 controller.go:1135] delete "pvc-7535d094-2366-11ea-a145-72982b3e8f81": persistentvolume deleted
I1220 20:22:43.117003       1 controller.go:1137] delete "pvc-7535d094-2366-11ea-a145-72982b3e8f81": succeeded
```

Implementation details of note:

- The Kubernetes nfs-provisioner is part of [kubernetes-incubator/external-storage](https://github.com/kubernetes-incubator/external-storage/tree/master/nfs)
- If you have a pre-existing NFS Server, use the [NFS Client Provisioner](https://github.com/helm/charts/tree/master/stable/nfs-client-provisioner) chart and provisioner instead.

### Azure File Setup

If one has a Kubernetes cluster on [Azure](https://azure.microsoft.com/en-us/services/kubernetes-service/),
and the `default` storageclass is of the non-`readWriteMany`-compatible `kubernetes.io/azure-disk` variety, one can create
an Azure File storageclass and then configure the Brigade project to use this instead of `default`.

See the official [Azure File storageclass example](https://kubernetes.io/docs/concepts/storage/storage-classes/#azure-file)
for the yaml to use.  _(Hint: The parameters section can be omitted altogether and Azure will use the defaults associated
with the existing Kubernetes cluster.)_

Create the resource via `kubectl create -f azure-file-storage-class.yaml`.

Finally, be sure to set the `buildStorageClass` and/or `cacheStorageCass` values to `azurefile` as above in the `nfs` example.

## Errata

- At this point, cache PVCs are never destroyed, even if the project to which
  they belong is destroyed. This behavior may change in the future.
- Killing the worker pod will orphan shared storage PVCs, as the cleanup routine
  is part of the worker's shutdown process. If you manually destroy a worker pod,
  you must also manually destroy the associated PVCs.
