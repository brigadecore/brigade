---
linkTitle: Deployment
title: Deploying to Production
description: How to configure and install a production-grade Brigade deployment
section: operators
weight: 1
aliases:
  - /deploy.md
  - /topics/deploy.md
  - /topics/operators/deploy.md
---

This guide will cover configuration and installation of a production-grade
Brigade deployment. If you are looking for a more comprehensive introduction to
Brigade or are interested in evaluating Brigade in a local, development-grade
cluster, view our [QuickStart](/intro/quickstart/) instead.

* [Prerequisites](#prerequisites)
* [Preparation](#preparation)
  * [Configure Host Name](#configure-host-name)
  * [Configure Passwords](#configure-passwords)
  * [Disable the Root User](#disable-the-root-user)
  * [Configure Ingress](#configure-ingress)
  * [Configure TLS](#configure-tls)
  * [Configure an Authentication Provider](#configure-an-authentication-provider)
  * [Configure MongoDB](#configure-mongodb)
  * [Configure ActiveMQ Artemis](#configure-activemq-artemis)
  * [Configure Shared Storage](#configure-shared-storage)
  * [Other Configuration Options](#other-configuration-options)
* [Install Brigade](#install-brigade)
* [Update DNS](#update-dns)
* [Verify the Deployment](#verify-the-deployment)
  * [Install the Brigade CLI](#install-the-brigade-cli)
  * [Log In](#log-in)
* [Wrap-Up](#wrap-up)
* [Deploying Multiple Brigades](#deploying-multiple-brigades)

## Prerequisites

* A remote, production-grade Kubernetes v1.16.0+ cluster. Your cluster must be
  capable of the following:
    * Provisioning public IPs. (If you're using a managed Kubernetes service
      from any prominent cloud provider, for instance, you're all set.)
    * Provisioning volumes that can be mounted to multiple pods simultaneously
      over the network. If your cloud provider or Kubernetes distro does not
      already provide a suitable
      [`StorageClass`]((https://kubernetes.io/docs/concepts/storage/storage-classes/))
       for this purpose, the
      [NFS Server Provisioner chart](https://github.com/helm/charts/tree/master/stable/nfs-server-provisioner),
      although deprecated, provides a viable and popular option.
* [Helm v3.7.0+](https://helm.sh/docs/intro/install/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* Free disk space on the cluster nodes. The installation requires sufficient
  free disk space and will fail if a cluster node disk is nearly full.

## Preparation

1. Enable Helm's experimental OCI support:

    **POSIX**
    ```shell
    $ export HELM_EXPERIMENTAL_OCI=1
    ```

    **PowerShell**
    ```powershell
    > $env:HELM_EXPERIMENTAL_OCI=1
    ```

1. Extract the default configuration from the Helm chart and save it to a
   convenient location. In the example below, we save it to
   `~/brigade-values.yaml`. We'll use that path to refer to that file throughout
   the remainder of this document, but if you save it to a different location,
   make the appropriate substitutions wherever you see that path.

```shell
$ helm inspect values oci://ghcr.io/brigadecore/brigade \
    --version v2.3.1 > ~/brigade-values.yaml
```

In the next steps, we'll edit `~/brigade-values.yaml` to configure a
production-grade deployment.

### Configure Host Name

In the Brigade chart values file (`~/brigade-values.yaml`), setting
`apiserver.host` has a default value of `localhost`. This is a suitable value
for a non-production deployment of Brigade, but for a variety of reasons, this
value should be set correctly for a production-grade deployment of Brigade.

Change the value of `apiserver.host` to reflect the DNS host name by which
end-users will log in via the CLI.

### Configure Passwords

In the Brigade chart values file (`~/brigade-values.yaml`) there is one
password with a hard-coded default. For production-grade deployments, it is
critical to supply your own value for that field.

You may _optionally_ supply your own values for the following fields as well:

* `apiserver.rootUser.password`: If you do not set a value for this, one will be
  generated on initial install and _will not_ change on subsequent
  `helm upgrade` operations unless explicitly overridden.
* `mongodb.auth.passwords`: If you do not set a value for this, one will be
  generated on initial install and _will be_ regenerated/changed by every
  subsequent `helm upgrade` operation unless explicitly overridden. Since you
  will rarely, if ever, need to use this password, this is generally not a
  problem. If you're personally in the habit of using a password manager and it can
  generate strong passwords for you, consider using that.
* `artemis.password`: If you do not set a value for this, one will be
  generated on initial install and _will not_ change on subsequent
  `helm upgrade` operations unless explicitly overridden.

### Disable the Root User

Production-grade Brigade deployments utilize third-party identity providers for
authentication. We'll configure this a little later.

Since Brigade's default configuration is optimized for getting started very
quickly in a local, development-grade cluster, the default configuration does
_not_ utilize any third-party identity provider for authentication. Instead, it
exposes a single, "root" user that has full administrative privileges.

We highly recommend _disabling_ the root user for production deployments. To do
so, set the value of of `apiserver.rootUser.enabled` to `false`.

## Configure Ingress

The term "ingress" can be overloaded. Although we _will_ discuss Kubernetes
[`Ingress` resources](https://kubernetes.io/docs/concepts/services-networking/ingress/)
below, this section is really about ingress in a more general sense -- how
traffic reaches Brigade's API server.

In the default configuration, the API server's Kubernetes `Service` is
configured as type `ClusterIP`. This means the API server does not receive a
public IP and will therefore not be reachable from outside your Kubernetes
cluster.

To make your API server reachable from outside your cluster (e.g. to users of
the `brig` CLI), you must do one of the following:

* **Enable Ingress:** (Here "ingress" refers to a Kubernetes `Ingress`
   resource.) Leave `apiserver.service.type` set to `CLusterIP` and change
   `apiserver.ingress.enabled` to `true`. If your
   [ingress controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/)
   requires you to decorate your `Ingress` resources with any
   controller-specific annotations, add those under
   `apiserver.ingress.annotations`.

* **Use a LoadBalancer:** Leave `apiserver.ingress.enabled` set to `false` and
  change `apiserver.service.type` to `LoadBalancer`.

### Configure TLS

In a production-grade Brigade deployment, the API server must _always_ secure
communication with the API server using TLS. If this is disabled, then the
secrecy of user credentials cannot be ensured as requests traverse unsecured
networks (e.g. the internet).

In the default configuration, self-signed TLS certificates are auto-generated
for the API server during installation. Because self-signed certificates are not
trusted, these should be overridden with valid, trusted certificates for
production-grade deployments.

How you proceed depends on choices made in the previous section:

* If you are using an ingress controller to route inbound traffic (see previous
  section), set `apiserver.ingress.tls.generateSelfSignedCert` to `false`. Once
  auto-generation of a cert is disabled, you become responsible for providing
  the certificate yourself. There are two ways to do this:

  * Provide cert/key material directly by (individually) base64-encoding your
    PEM encoded certificate and key and adding them, respectively, as the values
    for `apiserver.ingress.tls.cert` and `apiserver.ingress.tls.key`.

  * Alternatively, ensure the existence of a
    [Kubernetes cert `Secret`](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets)
    resource named `<Helm release name>-apiserver-ingres-cert` in the same
    namespace as Brigade. This would typically be
    `brigade-apiserver-ingres-cert` in the `brigade` namespace. _How_ you
    achieve this is up to you, but an easy and sensible way to accomplish it is
    through the use of a [cert manager].

* If you are _not_ using an ingress controller to route inbound traffic, and
  instead opted to use a `Service` of type `LoadBalancer` (see previous
  section), set `apiserver.tls.generateSelfSignedCert` to `false`. Once
  auto-generation of a cert is disabled, you become responsible for providing
  the certificate yourself. There are two ways to do this:

  * Provide cert/key material directly by (individually) base64-encoding your
    PEM encoded certificate and key and adding them, respectively, as the values
    for `apiserver.tls.cert` and `apiserver.tls.key`.

  * Alternatively, ensure the existence of a
    [Kubernetes cert `Secret`](https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets)
    resource named `<Helm release name>-apiserver-cert` in the same namespace as
    Brigade. This would typically be `brigade-apiserver-cert` in the `brigade`
    namespace. _How_ you achieve this is up to you, but an easy and sensible way
    to accomplish it is through the use of a [cert manager].

[Cert Manager]: https://cert-manager.io/docs/

## Configure an Authentication Provider

Production-grade Brigade deployments utilize third-party identity providers for
authentication. Any identity provider that supports
[OpenID Connect](https://openid.net/connect/) should work. Brigade has been
verified to work with
[Azure Active Directory](https://azure.microsoft.com/en-us/services/active-directory/)
and [Google Identity Platform](https://cloud.google.com/identity-platform).
[GitHub](https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps)
does not implement OpenID Connect, but is also supported due to its popularity.

* For identity providers that support OpenID Connect, set
  `apiserver.thirdPartyAuth.strategy` to `oidc`, then set each of the following
  to values provided by your identity provider:

  * `apiserver.thirdPartyAuth.oidc.providerURL`
  * `apiserver.thirdPartyAuth.oidc.clientID`
  * `apiserver.thirdPartyAuth.oidc.clientSecret`

* For GitHub, set `apiserver.thirdPartyAuth.strategy` to `github`, then set each
  of the following to to values provided by GitHub after setting up a
  [GitHub OAuth App](https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps):

  * `apiserver.thirdPartyAuth.github.clientID`
  * `apiserver.thirdPartyAuth.github.clientSecret`

  Optionally, enumerate GitHub organizations whose members are permitted to
  authenticate to your Brigade API server under
  `apiserver.thirdPartyAuth.github.allowedOrganizations`.

Regardless of which third-party identity provider you select, we recommend using
the `apiserver.thirdPartyAuth.admins` field to enumerate users who should
automatically be granted administrative privileges upon initial login. Without
this set, the root user would have to be temporarily enabled to facilitate
initial setup -- and we do not recommend that. Note that removing someone from
this list before performing a `helm upgrade` does _not_ revoke their existing
administrative privileges nor does adding a user to this list grant
administrative privileges if they have already logged in for the first time.
This list strictly manages auto-grant of administrative privileges on _first
login_ for each of the enumerated users.

### Configure MongoDB

[MongoDB](https://www.mongodb.com/) is used to persist the API server's user
data, project data, event data, logs, and more. If your use cases require
Brigade to be highly available or you have specific data retention requirements,
you may wish to tweak various aspects of the MongoDB deployment in the `mongodb`
section of `~/brigade-values.yaml`. Extensive discussion of this is out of scope
for this guide. Brigade utilizes
[MongoDB packaged by Bitnami](https://bitnami.com/stack/mongodb/helm) as a
sub-chart. Refer directly to their documentation for further details.

We _do_ strongly recommend increasing the size of volumes utilized by MongoDB
by setting `mongodb.persistence.size` to _at least_ 40 gigabytes (`40Gi`).

### Configure ActiveMQ Artemis

Brigade's event bus is implemented using
[ActiveMQ Artemis](https://activemq.apache.org/components/artemis/). If your
uses cases require Brigade to be highly available, you may wish to set
`artemis.ha.enabled` to `true`. This will establish a "warm," secondary node to
fail over to in the event that the primary node should fail. Failover and
fail back are automatic.

We do _not_ recommend modifying Artemis topology to use distributed queues
because it undermines the guarantee that events for a given project are handled
on a FIFO basis.

We _do_ strongly recommend increasing the size of volumes utilized by Artemis
by setting `artemis.persistence.size` to _at least_ 40 gigabytes (`40Gi`).

### Configure Shared Storage

Brigade workers and jobs may optionally mount a shared storage volume. This
provides a convenient mechanism for jobs to persist results or artifacts to a
location that's accessible to the worker and to downstream jobs.

Brigade is pre-configured to use Kubernetes
[`StorageClass`](https://kubernetes.io/docs/concepts/storage/storage-classes/)
`nfs` for dynamically shared storage volumes. If you'd like to use an
alternative `StorageClass`, override the value of
`worker.workspaceStorageClass`. The `StorageClass` used _must_ support access
mode `ReadWriteMany`.

### Other Configuration Options

Although we've covered the most critical, consider perusing
`~/brigade-values.yaml` further to discover other settings you may wish to fine
tune. The file itself is liberally commented with detailed instructions.

## Install Brigade

Finally, it's time. With `~/brigade-values.yaml` updated with configuration
suitable for production, we can proceed with installation:

```shell
$ helm install brigade \
    oci://ghcr.io/brigadecore/brigade \
    --version v2.3.1 \
    --create-namespace \
    --namespace brigade \
    --values ~/brigade-values.yaml \
    --wait \
    --timeout 300s
```

> ⚠️ Installation and initial startup may take a few minutes to complete.

## Update DNS

Now that Brigade is deployed, we'll determine the public IP address of the API
server or the ingress controller that will route inbound traffic _to_ the API
server. Which of these is applicable depends on the choice you made in the
[Configure Ingress](#configure-ingress) section.

* If you are _not_ using an ingress controller to route inbound traffic to your
  API server, use the following command to determine the API server's public IP:

  ```shell
  $ kubectl get svc brigade-apiserver --namespace brigade \
      --output jsonpath='{.status.loadBalancer.ingress[0].ip}'
  ```

* If you _are_ using an ingress controller to route inbound traffic to your API
  server, determine the public IP of that ingress controller. The procedure for
  this will vary from one ingress controller to the next, but will generally
  involve querying the ingress controller's `Service` of type `LoadBalancer`.
  The following example depicts how Brigade's maintainers make this
  determination for our own cluster's
  [Nginx Ingress Controller](https://kubernetes.github.io/ingress-nginx/):

  ```shell
  $ kubectl get svc nginx-ingress-nginx-controller --namespace nginx \
      --output jsonpath='{.status.loadBalancer.ingress[0].ip}'
  ```

With the public IP in hand, user your domain's DNS provider to create an
[A record](https://www.cloudflare.com/learning/dns/dns-records/dns-a-record/)
to map a hostname (the value of `apiserver.host`) to the public IP.

## Verify the Deployment

In this section, we'll verify our ability to at least log in to our new,
production-grade Brigade deployment.

### Install the Brigade CLI

If you haven't done so already, install the Brigade CLI, `brig`. In general, it
can be installed by downloading the appropriate pre-built binary from our
[releases page](https://github.com/brigadecore/brigade/releases) to a directory
on your machine that is included in your `PATH` environment variable. Below are
instructions for common environments:

**Linux**

```shell
$ curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-linux-amd64
$ chmod +x /usr/local/bin/brig
```

**macOS**

The popular [Homebrew](https://brew.sh/) package manager provides the most
convenient method of installing the Brigade CLI on a Mac:

```shell
$ brew install brigade-cli
```

Alternatively, you can install manually by directly downloading a pre-built
binary:

```shell
$ curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-darwin-amd64
$ chmod +x /usr/local/bin/brig
```

**Windows**

```powershell
> mkdir -force $env:USERPROFILE\bin
> (New-Object Net.WebClient).DownloadFile("https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-windows-amd64.exe", "$ENV:USERPROFILE\bin\brig.exe")
> $env:PATH+=";$env:USERPROFILE\bin"
```

The script above downloads `brig.exe` and adds it to your `PATH` for the current
session. Add the following line to your
[PowerShell Profile](https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/)
if you want to make the change permanent:

```powershell
> $env:PATH+=";$env:USERPROFILE\bin"
```

### Log In

Now you're ready to log in to Brigade! For the server URL value, use the DNS
hostname configured above. In the example below, we use `brigade.example.com`:

```shell
$ brig login --server https://brigade.example.com
```

This command will return a URL which can be copied, then pasted into your web
browser to complete authentication using the configured identity provider. We
recommend using the optional `--browse` (or `-b`) flag to bypass the manual
copy/paste process. Using that flag will immediately navigate to the
authentication URL using your system's default browser.

Upon successful authentication, you will be redirected to a splash page that
informs you that you may resume using the Brigade CLI.

## Wrap-Up

Now that you have a production-grade Brigade deployment, day-to-day operations
can all be completed using the `brig` CLI.

Upgrading Brigade's server-side components to a newer release or updating
Brigade's configuration can be accomplished with the `helm upgrade` command.
Uninstalling Brigade can be accomplished with `helm uninstall`. For more details
on using Helm, refer directly
to [Helm's own documentation](https://helm.sh/docs/intro/using_helm).

## Deploying Multiple Brigades

It is possible to deploy multiple Brigade instances to a single
Kubernetes cluster. However, there is one piece of configuration necessary for
each Brigade instance other than the original. This is due to the global RBAC
resources that are created by the original deployment.

For each subsequent Brigade deployment, set `rbac.installGlobalResources` to
`false`, then deploy as usual.
