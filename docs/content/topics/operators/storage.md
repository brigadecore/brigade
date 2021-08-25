---
title: Storage
description: Storage and configuration in Brigade
section: operators
weight: 3
aliases:
  - /storage.md
  - /intro/storage.md
  - /topics/storage.md
---

# Storage in Brigade

Brigade utilizes storage for its in the following ways:

  * [Shared Worker storage](#shared-worker-storage) wherein a Brigade Worker's
    workspace may be shared with its Job(s). Optional; not enabled by default.
  * [Artemis storage](#artemis-storage) for Brigade's Messaging/Queue component
  * [MongoDB storage](#mongodb-storage) for Brigade's backing data store

## Shared Worker storage

The workspace for a Brigade Worker can be shared with all Worker Jobs. This is
an opt-in feature and isn't enabled by default. When enabled, Brigade will
create a [PersistentVolume] on the underlying Kubernetes substrate and
automatically add the corresponding volume mount to each Worker Job created.

> Note: As this volume may be accessed by more than one pod, and each pod may
need both read and write access to the shared volume, its access mode is
[ReadWriteMany][Access Modes], which may not be supported by the default
[storage class] configured on your Kubernetes substrate. See the [Access Modes]
matrix for compatibility. Brigade is well-tested using [NFS] and [Azure File]
on [AKS]. ([Azure Disk] does *not* support this required access mode.)

To configure the storage class that should be used for shared Worker storage,
set the `worker.workspaceStorageClass` field in Brigade's [Helm chart values]
file to the name of the storage class. For example, if NFS is set up on the
Kubernetes cluster and its storage class is named `nfs`, this would be the
appropriate configuration:

```yaml
worker:
  workspaceStorageClass: nfs
```

[PersistentVolume]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/
[Access Modes]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes
[storage class]: https://kubernetes.io/docs/concepts/storage/storage-classes/
[NFS]: https://github.com/kubernetes-sigs/nfs-ganesha-server-and-external-provisioner/tree/master/deploy/helm
[Azure File]: https://azure.microsoft.com/en-us/services/storage/files/
[AKS]: https://azure.microsoft.com/en-us/services/kubernetes-service/
[Azure Disk]: https://azure.microsoft.com/en-us/services/storage/disks/
[Helm chart values]: https://github.com/brigadecore/brigade/blob/v2/charts/brigade/values.yaml

### Enabling Worker storage

To enable shared Worker storage, set `useWorkspace` to `true` under the
`workerTemplate` section on the [project configuration file][project file]
(usually `project.yaml`) for a Project. For example, here is the relevant bit
of configuration from the [10-shared-workspace example project]:

```yaml
spec:
  workerTemplate:
    useWorkspace: true
```

Each Worker Job requiring access to this workspace must then be configured with
a filepath value designating where the workspace should be mounted within the
Job's container. (Note that this may be the Job's `primaryContainer` and/or one
or more of a Job's `sidecarContainer`(s), if applicable). This filepath value
is assigned to the `workspaceMountPath` field on each applicable Job container.

In the example `brigade.js` script below, the Job named "first-job" is
configured with the `workspaceMountPath` value set to `/share` on both its
primary container and its sidecar container, "sidecar". The sidecar container
is able to read (and display) the message that the primary container prints,
via the shared workspace mount:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

events.on("brigade.sh/cli", "exec", async event => {
  let job = new Job("first-job", "debian:latest", event);
  job.primaryContainer.command = ["bash"];
  job.primaryContainer.arguments = ["-c", "echo 'Hello!' > /share/message"];
  job.primaryContainer.workspaceMountPath = "/share";

  job.sidecarContainers = {
    sidecar: {
      image: "debian:latest",
      command: ["cat"],
      arguments: ["/share/message"],
      workspaceMountPath: "/share"
    }
  };

  await job.run();
});

events.process();
```

[project file]: /topics/project-developers/projects#project-definition-files
[10-shared-workspace example project]: https://github.com/brigadecore/brigade/blob/v2/examples/10-shared-workspace/project.yaml

## Artemis storage

Brigade uses [ActiveMQ Artemis] as its messaging queue component. For more
details on its function in Brigade, see the [Design] doc.

Messages (i.e. work to be scheduled on the substrate) should be persisted even
when/if Artemis itself goes down or is restarted. Therefore, by default,
persistence is enabled via the appropriate configuration on the
[Brigade Helm Chart][Helm chart values]. The default access mode for the
backing PersistentVolume is `ReadWriteOnce`. This mode, although configurable,
should not need changing.

There are, however, other persistence options that may be useful to customize,
such as volume size and storage class type. All configuration options can be
seen under the `artemis.persistence` section of the Brigade chart.

As the access mode is `ReadWriteOnce`, nearly all storage class types should
function correctly for this PersistentVolume. By default, no storage class is
specified in the chart, which means the default storage class on the substrate
will be employed.

[Design]: /topics/design
[ActiveMQ Artemis]: https://activemq.apache.org/components/artemis/

## MongoDB storage

Brigade uses [MongoDB] as its backing data store. For more details on its
function in Brigade, see the [Design] doc.

Data should be persisted even when/if MongoDB itself goes down or is restarted.
Therefore, by default, persistence is enabled via the appropriate configuration
on the [Brigade Helm Chart][Helm chart values]. The default access mode for the
backing PersistentVolume is `ReadWriteOnce`. This mode, although configurable,
should not need changing.

There are, however, other persistence options that may be useful to customize,
such as volume size and storage class type. All configuration options can be
seen under the `mongodb.persistence` section of the Brigade chart.

As the access mode is `ReadWriteOnce`, nearly all storage class types should
function correctly for this PersistentVolume. By default, no storage class is
specified in the chart, which means the default storage class on the substrate
will be employed.

[MongoDB]: https://www.mongodb.com/

## Examples

### Using an NFS Server

For Brigade Worker storage, it is often convenient to use storage backends that
are optimized for short-term ephemeral storage.

NFS (Network File System) is one protocol that works well for Brigade. You can
use the [NFS Server Provisioner][NFS] chart to easily install an NFS server.

```console
$ helm repo add stable ttps://charts.helm.sh/stable
$ helm install nfs stable/nfs-server-provisioner \
  --create-namespace --namespace nfs
```

By default, the chart installs with persistance disabled. For various methods
on enabling, as well as configuring other aspects of the installation, see the
[README][NFS].

This chart installs a [StorageClass][storage class] named `nfs`. To configure
Brigade to use this storage class for shared Worker storage, set
`worker.storageClass` to `nfs` in the
[Brigade Helm chart values file][Helm chart values] and supply this file on
install/upgrade.
