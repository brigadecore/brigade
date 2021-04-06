# Brigade 2: Event-based Scripting for Kubernetes

Brigade 2 is currently in an _alpha_ state and remains under heavy development.

## Introducing Brigade 2

Brigade 2 has been lovingly re-engineered from the ground up. We believe we've
remained faithful to the original vision of Brigade 1.x, and as such, much
general knowledge of Brigade 1.x can be carried over.

_But we've also learned a lot from Brigade 1.x._ Brigade 2 has been designed,
_explicitly_ to reduce the degree of Kubernetes knowledge required for success.
While Brigade 1.x was hailed as "event driven scripting for Kubernetes," Brigade
2 is "event driven scripting (for Kubernetes)." Moreover, great care has been
taken to improve security and scalability, and with our all new API and
complementary SDKs, we're also lowering barriers to integration.

We hope you'll enjoy this product refresh as much as we are.

## Getting Started

Comprehensive documentation will become available in conjunction with a beta
release. In the meantime, here is a little to get you started.

### Installing Brigade 2 on a _Private_ Kubernetes Cluster

For now, we're using the [GitHub Container Registry](https://ghcr.io) (which is
an [OCI registry](https://helm.sh/docs/topics/registries/)) to host our Helm
chart. Helm 3 has _experimental_ support for OCI registries. In the event that
the Helm 3 dependency proves troublesome for Brigade users, or in the event that
this experimental feature goes away, or isn't working like we'd hope, we will
revisit this choice before going GA.

To install Brigade 2 with _default_ configuration:

```console
$ export HELM_EXPERIMENTAL_OCI=1
$ helm chart pull ghcr.io/brigadecore/brigade:v2.0.0-alpha.2
$ helm chart export ghcr.io/brigadecore/brigade:v2.0.0-alpha.2 -d ~/charts
$ kubectl create namespace brigade2
$ helm install brigade2 ~/charts/brigade --namespace brigade2
```

__Please take note that the default configuration is not secure and is not
appropriate for _any_ shared cluster. This is on account of hardcoded passwords,
auto-generated, self-signed certificates, and the enablement of Brigade's "root"
user. _This configuration is appropriate for evaluating Brigade 2 in a private
cluster only_ (for instance, a local
[minikube](https://minikube.sigs.k8s.io/docs/) or
[kind](https://kind.sigs.k8s.io/) cluster, or any cluster used exclusively by
oneself).__

To view configuration options:

```console
$ helm inspect values ~/charts/brigade > ~/brigade2-values.yaml
```

To apply alternative configuration, edit `~/brigade2-values.yaml` as you see
fit, then:

```console
$ helm install brigade2 ~/charts/brigade --namespace brigade2
```

### Exposing the Brigade 2 API Server

Because you are presumably following these steps in a local cluster, the best
method of exposing Brigade 2's API server is to do something like this after
installation:

```console
$ kubectl --namespace brigade2 port-forward service/brigade2-apiserver 8443:443 &>/dev/null &
```

### Installing the `brig` CLI

Next, download the appropriate, pre-built `brig` CLI (command line interface)
from our [releases page](https://github.com/brigadecore/brigade/releases) and
move it to any location on your path, such as `/usr/local/bin`, for instance.

### Logging In

Log in as the "root" user, using the default root password `F00Bar!!!`. Be sure
to use the `-k` option to disregard issues with the self-signed certificate.

```console
$ brig -k login --server https://localhost:8443 --root
```

For security reasons, root user sessions are invalidated one hour after they
are created. If you play with Brigade 2 for more than an hour, or you walk away
and come back, you will have to log in again.

For drastically improved security, we support authentication using [Open ID
Connect](https://openid.net/connect/) and third-party identity providers like
[Azure Active
Directory](https://azure.microsoft.com/en-us/services/active-directory/) or
[Google Cloud Identity Platform](https://cloud.google.com/identity-platform/),
but configuring that is a bit more involved and doesn't work well if you're
taking Brigade 2 for a test drive in a local environment like minikube or kind.

### Creating a Project

Your next step is to create a Brigade __project__. Unlike Brigade 1.x, this is
not accomplished by means of an onerous, interactive process. Rather, it is
accomplished using a file that looks suspiciously like a Kubernetes manifest
(but isn't).

You can download an example from
[here](https://raw.githubusercontent.com/brigadecore/brigade/v2/examples/javascript/pipeline-demo.yaml):

With this file stored locally, at a location such as `~/pipeline-demo.yaml`, for
instance, you can direct Brigade to create a new project from this file:

```console
$ brig -k project create --file ~/pipeline-demo.yaml
```

If you want to alter the example, note that with an appropriate editor or IDE
(we use [VS Code](https://code.visualstudio.com/)) and appropriate plugins (we
use [this
one](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)),
you can receive context help while editing the example!

### Creating an Event

With your first project set up, it's time to create your first event:

```console
$ brig -k event create --project pipeline-demo
```

On success, this step will reveal the ID of the new event, which will be handled
_asynchronously_ by Brigade 2.

### Watch the Event

To view the status of the event:

```console
$ brig -k event get --id <event id from previous step>
```

Eventually, the worker spawned to process the event, and any jobs spawned by
the worker, should all display a `SUCCEEDED` status.

Congratulations! You're using Brigade 2!

## The Brigade 2 Ecosystem

Brigade 2's all new API lowers the bar to creating all manner of peripherals--
tooling, event gateways, and more. To facilitate this, we have three
complementary SDKs in varying states of development and maturity that should
make it a snap to create interesting things that interact with Brigade 2.

Be sure to check out:

* [Brigade SDK for Go](https://github.com/brigadecore/brigade/tree/v2/sdk) (used by Brigade itself)
* [Brigade SDK for JavaScript](https://github.com/krancour/brigade-sdk-for-js) (and TypeScript)]
* [Brigade SDK for Rust](https://github.com/brigadecore/brigade-sdk-for-rust) (still very new)

## More Docs to Come

Stay tuned for complete and comprehensive developer and operator documentation
in a future release.

## Contributing

The Brigade project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Support & Feedback

We have a slack channel!
[Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free
to join for any support questions or feedback, we are happy to help. To report
an issue or to request a feature open an issue
[here](https://github.com/brigadecore/brigade/issues)
