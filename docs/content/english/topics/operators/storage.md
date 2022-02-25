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

Brigade utilizes storage in the following ways:

  * [Shared Worker storage](#shared-worker-storage) wherein a Brigade Worker's
    workspace may be shared with and among its Jobs
  * [Artemis storage](#artemis-storage) for Brigade's Messaging/Queue component
  * [MongoDB storage](#mongodb-storage) for Brigade's backing data store

## Shared Worker storage

The workspace for a Brigade Worker can be shared among all Worker Jobs. This is
an opt-in feature and isn't enabled by default. When enabled, Brigade will
create a [PersistentVolume] on the underlying Kubernetes cluster and
automatically add the corresponding volume mount to each Worker Job created.

> Note: As this volume may be accessed by more than one pod, and each pod may
need both read and write access to the shared volume, its access mode is
[ReadWriteMany][Access Modes], which may not be supported by the default
[storage class] configured on your Kubernetes cluster. See the [Access Modes]
matrix for compatibility. Brigade is well-tested using [NFS] and [Azure File]
on [AKS]. ([Azure Disk] does *not* support this required access mode.)

For Brigade Worker storage, it is often convenient to use storage backends that
are optimized for short-term ephemeral storage. To that end, Brigade ships with
its default configured as NFS (Network File System). Therefore, NFS will need
to be deployed on the same Kubernetes cluster as Brigade. You can use the
[NFS Server Provisioner][NFS] chart for this purpose:

```shell
$ helm repo add stable ttps://charts.helm.sh/stable
$ helm install nfs stable/nfs-server-provisioner \
  --create-namespace --namespace nfs
```

By default, the NFS chart installs with persistance disabled. For various
methods on enabling, as well as configuring other aspects of the installation,
see the [README][NFS].

This chart installs a [StorageClass][storage class] named `nfs`. As mentioned,
Brigade already has `worker.storageClass` set to `nfs` in its
[Helm chart values file][Helm chart values]. To configure an alternate storage
class, set this field's value to the preferred storage class name. For example,
to use the Azure File storage class, the appropriate configuration would be:

```yaml
worker:
  workspaceStorageClass: azurefile
```

[PersistentVolume]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/
[Access Modes]: https://kubernetes.io/docs/concepts/storage/persistent-volumes/#access-modes
[storage class]: https://kubernetes.io/docs/concepts/storage/storage-classes/
[NFS]: https://github.com/kubernetes-sigs/nfs-ganesha-server-and-external-provisioner/tree/master/deploy/helm
[Azure File]: https://azure.microsoft.com/en-us/services/storage/files/
[AKS]: https://azure.microsoft.com/en-us/services/kubernetes-service/
[Azure Disk]: https://azure.microsoft.com/en-us/services/storage/disks/
[Helm chart values]: https://github.com/brigadecore/brigade/blob/main/charts/brigade/values.yaml

### Enabling Worker storage

To enable shared Worker storage, set `useWorkspace` to `true` under the
`workerTemplate` section on the [project configuration file][project file]
(usually `project.yaml`) for a Project. For example, here is the relevant bit
of configuration from the [08-shared-workspace example project]:

```yaml
spec:
  workerTemplate:
    useWorkspace: true
```

Each Worker Job requiring access to this workspace must then be configured with
a filepath value designating where the workspace should be mounted within the
Job's container. (Note that this may be the Job's `primaryContainer` and/or one
or more of a Job's `sidecarContainer`(s)). This filepath value is assigned to
the `workspaceMountPath` field on each applicable Job container.

In the example `brigade.js` script below, the both Jobs are configured with the
`workspaceMountPath` value set to `/share`. Thus, `second-job` is able to read
(and display) the message that `first-job` emits, via the shared workspace
mount:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

events.on("brigade.sh/cli", "exec", async event => {
  let job1 = new Job("first-job", "debian:latest", event);
  job1.primaryContainer.workspaceMountPath = "/share";
  job1.primaryContainer.command = ["bash"];
  job1.primaryContainer.arguments = ["-c", "echo 'Hello!' > /share/message"];
  await job1.run();

  let job2 = new Job("second-job", "debian:latest", event);
  job2.primaryContainer.workspaceMountPath = "/share";
  job2.primaryContainer.command = ["cat"];
  job2.primaryContainer.arguments = ["/share/message"];
  await job2.run();
});

events.process();
```

[project file]: /topics/project-developers/projects#project-definition-files
[08-shared-workspace example project]: https://github.com/brigadecore/brigade/blob/main/examples/08-shared-workspace/project.yaml

## Artemis storage

Brigade uses [ActiveMQ Artemis] as its messaging queue component. For more
details on its function in Brigade, see the [Design] doc.

Messages (i.e. work to be scheduled on Kubernetes) should be persisted even
if/when Artemis itself goes down or is restarted. Therefore, by default,
persistence is enabled via the appropriate configuration on the
[Brigade Helm Chart][Helm chart values]. The default access mode for the
backing PersistentVolume is `ReadWriteOnce`. This mode, although configurable,
should not need changing.

There are, however, other persistence options that may be useful to customize,
such as volume size and storage class type. Regarding volume size, the default
is a fairly small `8Gi`, which is not considered adequate for a production
deployment and should be updated accordingly. All configuration options can be
seen under the `artemis.persistence` section of the Brigade chart.

As the access mode is `ReadWriteOnce`, nearly all storage class types should
function correctly for this PersistentVolume. By default, no storage class is
specified in the chart, which means the default storage class on the Kubernetes
cluster will be employed.

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
such as volume size and storage class type. Regarding volume size, the default
is a fairly small `8Gi`, which is not considered adequate for a production
deployment and should be updated accordingly. All configuration options can be
seen under the `artemis.persistence` section of the Brigade chart.

As the access mode is `ReadWriteOnce`, nearly all storage class types should
function correctly for this PersistentVolume. By default, no storage class is
specified in the chart, which means the default storage class on the Kubernetes
cluster will be employed.

[MongoDB]: https://www.mongodb.com/
