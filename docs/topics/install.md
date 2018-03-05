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

This example shows how to check out the latest Brigade release:

```console
$ git clone https://github.com/azure/brigade.git
$ cd brigade
$ git checkout $(git tag -l | tail -n 1)
$ helm install ./charts/brigade
```

(If you are interested in building from `master`, see the [developer docs](developer.md).)

Once you have Brigade installed, you can proceed to [creating a project](projects.md).
The remainder of this guide covers special configurations of Brigade.

> If you are not working off of a tagged release, you may also have to build
> custom images. The [Developers Guide](developers.md) explains this in more
> detail. Otherwise, the images referenced by the chart will be from the last
> release, and may not have the latest changes.

### Customizing Installations

Both of these options use the latest Brigade release from DockerHub. But you can override
this behavior by supplying alternate images during installation.

For each component of Brigade, you can set it's image and tag separately:

```console
$ helm install brigade/brigade --set controller.name=my-image --set controller.tag=1.2.3
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

The following volume backends have been tested:

- Azure: AzureFile
- Minikube: 9p (the default)
- Any: NFS

The NFS solution is explained below.

#### Using an NFS Server

Brigade uses storage for caching and short-term file sharing. Because of this,
it is often convenient to use storage backends that are optimized for short-term
ephemeral storage.

NFS (Network File System) is one protocol that works well for Brigade. You can
use the [NFS Provisioner](https://github.com/IlyaSemenov/nfs-provisioner-chart)
chart to easily install an NFS server.

```console
$ helm repo add nfs-provisioner https://raw.githubusercontent.com/IlyaSemenov/nfs-provisioner-chart/master/repo
$ helm install --name nfs nfs-provisioner/nfs-provisioner --set hostPath=/var/run/nfs-provisioner
```

(Note that RBAC is enabled by default. To turn it off, use `--set rbac.enabled=false`.)

To use an emptyDir instead of a host mount, set `--hostPath=""`.


```console
$ helm repo add nfs-provisioner https://raw.githubusercontent.com/IlyaSemenov/nfs-provisioner-chart/master/repo
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


