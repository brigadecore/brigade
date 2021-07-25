---
title: Deployment
description: How to deploy and manage Brigade
section: operators
weight: 1
aliases:
  - /deploy.md
  - /topics/deploy.md
  - /topics/operators/deploy.md
---

# Deploying and managing Brigade

In this doc, we'll go over steps for setting up and managing a
production-grade Brigade deployment.  This includes:

  - [Configuring passwords](#configuring-passwords)
  - [Certificates and TLS](#certificates-and-tls)
  - [Managing external access to Brigade via an ingress resource](#managing-inbound-traffic-via-ingress)
  - [Choosing a third-party authentication provider for managing user auth](#third-party-authentication)

(Alternatively, if you're interested in trying Brigade in a private and/or local development
environment, take a look at the [QuickStart].)

[QuickStart]: /intro/quickstart.md

## Prerequisites

* A [Kubernetes cluster], version 1.16+.
  Your cluster must be accessible to the outside world, e.g. to source(s) of
  event triggers (gateways) and to browsers for authentication callbacks.
  As long as the cluster is able to provision a public ip address, as any
  cloud provider's Kubernetes offering will support, you should be good to go.
* [Helm] CLI v3 installed.
* [kubectl] CLI installed.
* Free disk space on the cluster nodes.
  The installation requires sufficient free disk space and will fail if a
  cluster node disk is nearly full.

[Kubernetes cluster]: https://kubernetes.io/docs/setup/
[Helm]: https://helm.sh/docs/intro/install/
[kubectl]: https://kubernetes.io/docs/tasks/tools/#kubectl

## Prepping the Brigade 2 installation

For now, we're using the [GitHub Container Registry](https://ghcr.io) (which is
an [OCI registry](https://helm.sh/docs/topics/registries/)) to host our Helm
chart. Helm 3 has _experimental_ support for OCI registries. In the event that
the Helm 3 dependency proves troublesome for Brigade users, or in the event that
this experimental feature goes away, or isn't working like we'd hope, we will
revisit this choice before going GA.

First, let's set the necessary environment variable to enable this experimental
support and pull the Brigade chart to a local directory:

```console
$ export HELM_EXPERIMENTAL_OCI=1
$ helm chart pull ghcr.io/brigadecore/brigade:v2.0.0-beta.1
$ helm chart export ghcr.io/brigadecore/brigade:v2.0.0-beta.1 -d ~/charts
```

We're now ready to view and edit configuration options:

```console
$ helm inspect values ~/charts/brigade > ~/brigade2-values.yaml
```

In the next steps, we'll go through the configuration needed for a
production-grade deployment.

## Configuring passwords

In the Brigade chart values file (`brigade2-values.yaml`) there are a few spots
where default passwords are used. It is recommended to supply your own values
for these fields. Here are the values to update and their locations in the
file:

  - The root user password at `apiserver.rootUser.password`
  - The MongoDB root user and database passwords at `mongodb.auth.rootPassword`
    and `mongodb.auth.password`
  - The Artemis user password at `artemis.password`

## Managing inbound traffic via Ingress

Before configuring Brigade to create an [Ingress resource] for its API server,
we need to deploy an [Ingress Controller] as well as create a DNS hostname
record which will map to the Ingress controller's external IP address.

[Ingress resource]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/

### Deploying the Ingress Controller

There are many Ingress Controller options.  For this guide, we will use the
[NGINX Ingress Controller].  Installation is as simple as adding the
corresponding chart repo and installing the chart.  We'll use all of the
defaults that ship with the chart, but full configuration can be explored via
the [chart's GitHub repository]

```console
$ helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
$ helm repo update
$ helm install ingress-nginx ingress-nginx/ingress-nginx
```

We need to acquire the external IP address provisioned for the NGINX Ingress
Controller's service.  We'll use this in the DNS record creation section next.

```console
$ kubectl -n kube-system get svc nginx-ingress-ingress-nginx-controller
```

It may take a few minutes for the address to be provisioned.  Once provisioned,
capture the value under the `EXTERNAL-IP` header.

In Brigade's chart values file, under the `apiserver.service` section, set
`type` to `ClusterIP`.  Here is a condensed view with other nearby fields
omitted:

```yaml
apiserver:
  service:
    type: ClusterIP
```

In doing so, Brigade's own API server is only reachable internally to the
Kubernetes cluster.  All inbound traffic will be handled by the NGINX Ingress
Controller.

[chart's GitHub repository]: https://github.com/kubernetes/ingress-nginx/tree/main/charts/ingress-nginx
[NGINX Ingress Controller]: https://kubernetes.github.io/ingress-nginx/

### DNS hostname

With the external IP address of the ingress controller in hand, you're now
ready to create the DNS entry.  With your DNS provider of choice, you'll create
an [A record] mapping the the hostname on a domain you control to this IP
address.  For example, if you control the `example.com` domain and wish for the
hostname of the Brigade server to be `mybrigade.example.com`, create the DNS
record with `mybrigade` as the name, `A` as the record type and the
external IP address captured previously as the entry in the corresponding IP
addresses section.

[A record]: https://www.cloudflare.com/learning/dns/dns-records/dns-a-record/

### Adding Ingress configuration

Now we're ready to add configuration for ingress to the chart values file for
Brigade.  We'll also update this configuration after the next
[Certificates and TLS](#certificates-and-tls) section.

All configuration will go under the `apiserver` section of the
`brigade-values2.yaml` that you've saved locally above.  Here's the breakdown,
with other nearby sections omitted:

```yaml
apiserver:

  ## This is the DNS hostname you've reserved
  host: mybrigade.example.com

  ## Ensure tls is enabled for all Brigade's internal components
  ## For now, we'll have them use self-signed certs
  tls:
    enabled: true
    generateSelfSignedCert: true
    # cert: base 64 encoded cert goes here
    # key: base 64 encoded key goes here

  ## Here's where we configure ingress
  ingress:
    enabled: true
    ## These are annotations specific to the ingress controller we chose,
    ## the NGINX Ingress Controller
    annotations:
      kubernetes.io/ingress.class: nginx
      nginx.ingress.kubernetes.io/secure-backends: "true"
      nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
    tls:
      enabled: true
      ## We set this to false as we'll use an external certificate provider
      ## in the next step
      generateSelfSignedCert: false
```

Save these changes to `brigade-values2.yaml` and proceed to the next section.

## Certificates and TLS

To secure inbound requests to Brigade's API server and to avoid cert warnings
on the client's end (which occur when using the default self-signed certs), it
is recommended to select a cert issuer which uses a well-known CA.  (You may
also generate certs yourself, using a CA known to all clients, but this doc
will focus on using a third-party cert issuer.)

We'll select [Cert Manager] to serve as the certificate manager, [ACME]
as the issuer type and [Let's Encrypt] as the certificate source.

Cert Manager will create the Kubernetes resources associated with the ingress
hostname and manage their lifecycle, i.e. monitoring their validity and
attempting to renew before they expire.

Let's Encrypt will provision the certificate following the ACME protocol.

There will be three steps to set up certificate provisioning:

  - [Deploy Cert Manager](#deploy-cert-manager)
  - [Create Cert Issuers](#create-cert-issuers)
  - [Configure Brigade](#configure-brigade)

[Cert Manager]: https://cert-manager.io/docs/
[ACME]: https://cert-manager.io/docs/configuration/acme/
[Let's Encrypt]: https://letsencrypt.org/

### Deploy Cert Manager

To continue the theme of deploying these auxiliary services to Kubernetes via
Helm, we'll install Cert Manager via its Helm chart:

```console
$ helm repo add jetstack https://charts.jetstack.io
$ helm repo update
$ helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.4.0 \
  --set installCRDs=true
```

_Note: This is where the Kubernetes 1.18.8+ and Helm v3.3.1+ constraints come
in.  For more info, see [Cert Manager's upgrade notes]_

[Cert Manager's upgrade notes]: https://cert-manager.io/docs/installation/upgrading/upgrading-0.15-0.16/#helm

### Create Cert Issuers

We now need to configure the [ACME Cert Issuer resource] and create it via
`kubectl`.  We'll only create a production resource, but you may also create
a staging resource in addition or instead.  See the Cert Manager docs for more
info.

The prod resource (to be saved as `cert-manager-issuer-prod.yaml`) looks like
this:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # The ACME server URL
    server: https://acme-v02.api.letsencrypt.org/directory
    # Email address used for ACME registration
    email: youremail@goeshere.com
    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    # An empty 'selector' means that this solver matches all domains
    - selector: {}
      http01:
        ingress:
          class: nginx
```

The only field that is strictly necessary to update is the `email` field.

Once this file is updated and saved, we'll create it via kubectl:

```console
$ kubectl apply -f cert-manager-issuer-prod.yaml
```

[ACME Cert Issuer resources]: https://cert-manager.io/docs/configuration/acme/#creating-a-basic-acme-issuer

### Configure Brigade

Now we're ready to add cert manager configuration to the Brigade chart values
file, `brigade2-values.yaml`.  In fact, there are only a few lines to add under
the `apiserver.ingress.annotations` section (the same section updated when
configuring ingress):

```yaml
apiserver:
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: nginx
      # The next three lines represent the newly added configuration
      kubernetes.io/tls-acme: "true"
      cert-manager.io/cluster-issuer: "letsencrypt-prod"
      cert-manager.io/acme-challenge-type: http01
      nginx.ingress.kubernetes.io/secure-backends: "true"
      nginx.ingress.kubernetes.io/backend-protocol: "HTTPS"
```

## Third-Party Authentication

The default mode of authentication into Brigade, that of a root user and
password, is not suitable for a setup involving multiple users.

For production-grade user authentication, Brigade ships with support for
OIDC-compatible providers (such as [Google Identity Platform] and
[Azure Active Directory]) or [GitHub].  In this doc, we'll demonstrate
configuring GitHub to be our auth provider.

[Google Identity Platform]: https://cloud.google.com/identity-platform
[Azure Active Directory]: https://azure.microsoft.com/en-us/services/active-directory/
[GitHub]: https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps

### Configuring GitHub Authentication

To set up GitHub authentication with Brigade, we'll create a [GitHub OAuth App]
with the authorization callback URL set to point to the
`/v2/session/auth` endpoint on Brigade's API server address, e.g. the hostname
reserved when setting up ingress.  We'll then use the GitHub OAuth Apps's
client ID and a generated client secret in the `brigade2-values.yaml` file.

  1. Follow the [GitHub OAuth App] creation instructions, supplying your choice
    of values for `Application Name`, `Homepage URL` and `Application description`.
  1. For the `Authorization callback URL`, you'll supply the a value based
    on the DNS hostname and path mentioned above.  For example, if the DNS
    hostname is `mybrigade.example.com`, the value would be:
      ```
      https://mybrigade.example.com/v2/session/auth
      ```
  1. Click 'Register application'
  1. On the App settings page, there is now a Client ID string.  We'll use this
    value soon.
  1. Under `Client secrets`, click `Generate a new client secret`
  1. Save this client secret value now, as it won't be displayed again
  1. Click 'Update application'

Now that we have values for the GitHub OAuth App's client ID and client secret,
we're ready to update our Brigade chart values file with this auth
configuration.

All of it goes the `apiserver.thirdPartyAuth` section, which we show here,
with other nearby sections omitted:

```yaml
apiserver:
  ## Options for authenticating via a third-party authentication provider.
  thirdPartyAuth:
    ## Here we enter our chosen strategy of github
    strategy: github
    ## Here we inject the values from the GitHub OAuth App
    github:
      ## The client ID goes here
      clientID: foo
      ## The client secret goes here
      clientSecret: foo
      ## If only users from specific GitHub organizations should be allowed
      ## to authenticate, list them here.  Otherwise, users from any GitHub
      ## organization may attempt to authenticate.
      allowedOrganizations:
    ## User Session TTL dictates the default time-to-live for user sessions.
    ## Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    ## For example, "60s", "2h45m", "168h" (1 week)
    userSessionTTL: 168h
    ## If there are users that should be granted admin privileges in Brigade
    ## from the moment they authenticate, they should be listed here.  For
    ## instance, the operator doing the deployment and/or the user who will
    ## be in charge of assigning authorization permissions to authenticated
    ## users.
    ## Otherwise, permissions will need to be manually configured for each
    ## authenticated user.
    admins:
      - <enter your name here>
```

Don't forget to add your GitHub username under the `admins` list, so that you
as the operator have full admin privileges upon first login to Brigade.  Note:
since these privileges are granted on first login, adding/revoking these
permissions must be done via the brig CLI directly rather than re-deploying
with differing configuration.

Save the updated file and proceed to the next section.

[GitHub OAuth App]: https://docs.github.com/en/developers/apps/creating-an-oauth-app

## Additional configuration to consider

### Increasing volume size

Data volumes are used by both the backing data store and messaging queue
components, with implementations currently provided by MongoDB and Artemis,
respectively. For production deployments, it is recommended to increase the
size of both volumes.  As an example, Brigade's own CI/CD cluster is configured
with both volume sizes set to `40Gi`.

These values can be updated via the following two locations:

  - `mongodb.persistence.size`
  - `artemis.persistence.size`

## Deployment time!

Now that we have our `brigade2-values.yaml` updated with configuration around
ingress, TLS and third-party auth, you're ready to deploy!  Issue the following
`helm install` command, supplying a dedicated Kubernetes namespace where all
resources will be created and the filepath to the chart values file you've
saved: 

```console
$ helm install brigade2 ~/charts/brigade \
  --create-namespace
  --namespace brigade2 \
  --values ~/brigade2-values.yaml
```

You can issue the following `kubectl` command to monitor the status of the
deployments:

```console
$ kubectl rollout status deployment brigade2-apiserver -n brigade2 --timeout 5m
```

## Verifying deployment

Once all of the Kubernetes resources associated with Brigade are up and
running, you're ready to interact with Brigade.  We'll install the brig CLI and
verify TLS/cert setup, ingress and third-party auth, all via one command,
`brig login`.

### Installing the `brig` CLI

Next, download the appropriate, pre-built `brig` CLI (command line interface)
from our [releases page](https://github.com/brigadecore/brigade/releases) and
move it to any location on your path, such as `/usr/local/bin`, for instance.

### Logging In

Now you're ready to log in to Brigade!  For the server URL value, we use the
hostname configured above.  Say the hostname is `mybrigade.example.com`, the
login command would then be:

```console
$ brig login --server https://mybrigade.example.com
```

This command will return a GitHub Oauth URL which you'll paste into the
browser of your choice to complete authentication.

_Note: you can also supply the `-b/--browse` flag if you'd like the default
browser to automatically open a page to the authentication callback URL
returned by the command._

On first login, the browser will navigate to a GitHub webpage requesting access
on behalf of the GitHub Oauth App to high-level, read-only access to your
GitHub user account (basically, to read your username and organization
associations).  Once approved, you should see a Brigade webpage confirming that
you are now logged in and ready to use the Brigade CLI.

## Wrap-up

Now that you have a production-ready Brigade server deployed using established
services for managing traffic, certificates and authentication, you're ready to
accept new users and transition focus to the Brigade system itself.

As all of the services, including Brigade itself, have been deployed by their
corresponding Helm charts, lifecycle management and configuration changes are
as easy as executing a `helm upgrade` command, either to bump to a more recent
chart version or to update release configuration.  For more details around
managing Helm releases, see [Helm's documentation].

[Helm's documentation]: https://helm.sh/docs/intro/using_helm
