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
> thus share the cache. This behavior is EXPERIMENTAL and may be changed.

That PVC is never removed by Brigade. Each subsequent run of the same Job will
then mount that same PVC.

### Shared Storage

Shared storage provisioning is markedly different than caches.

- The worker will _always_ provision a shared storage PVC _per build_. (This
  behavior is experimental, and may change).
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
or that you're able to use [NFS](https://github.com/Azure/brigade/blob/master/docs/topics/install.md#using-an-nfs-server).

Only the following volume drivers are tested:

- Minikube's 9P implementation
- Azure's AzureFile storage
- NFS

We believe Gluster will work, but it's untested.

## Errata

- At this point, cache PVCs are never destroyed, even if the project to which
  they belong is destroyed. This behavior may change in the future.
- Killing the worker pod will orphan shared storage PVCs, as the cleanup routine
  is part of the worker's shutdown process. If you manually destroy a worker pod,
  you must also manually destroy the associated PVCs.
