---
title: Securing Brigade
description: How to configure security for Brigade.
section: operators
weight: 2
aliases:
  - /security.md
  - /intro/security.md
  - /topics/security.md
---

# Securing Brigade

The execution of Brigade scripts involves dynamically creating (and destroying)
a number of Kubernetes objects, including pods, secrets, and persistent volume
claims. For that reason, it is prudent to configure security.

- *Isolate Brigade in a namespace*: It is best to run Brigade in its own
  namespace. For example, when installing Brigade via its [Helm chart], do
  `helm install --namespace brigade ...`.
- *Multiple tenants in Brigade*: Brigade supports multiple projects per Brigade
  server instance, but it must be mentioned that users within Brigade can
  usually see (read) all projects on that instance, though they might not
  necessarily have the role necessary to write to them. If this presents a
  concern, each tenant should have its own Brigade instance. For more info,
  see the [Authentication] document.
- *Events should hold no sensitive data*: Because Brigade routes events to
  interested parties (projects) based on a subscription model, events should
  never contain secrets/sensitive information. Always assume that anyone in
  your cluster could be subscribed to any event that a gateway creates.

[Authentication]: /topics/administrators/authentication
  
## API Server Security

If Brigade's API server will be exposed to the internet either via a service of
type LoadBalancer having a public IP or via an [ingress controller], care
should be taken to secure inbound communication. Minimally, TLS should be
enabled and SSL certificates should not be self-signed.

For more details on deploying Brigade securely, see the [Deployment] doc.

[Helm chart]: https://github.com/brigadecore/brigade/tree/main/charts/brigade
[ingress controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
[Deployment]: /topics/operators/deploy

## How RBAC Is Configured

Brigade requires and assumes that the underlying Kubernetes cluster is
RBAC-enabled. Without RBAC, the risk that user-defined workloads may
accidentally or maliciously manipulate the Kubernetes cluster itself is high, so
such a configuration is strictly not supported.

The [Helm chart] for Brigade creates numerous RBAC-related resources that are
necessary for Brigade to function properly. These include `ServiceAccount`s and
applicable `ClusterRoleBinding`s for each of Brigade's components (API Server,
Observer, Scheduler, etc.) By default, the `ClusterRole`s referenced by those
`ClusterRoleBinding`s are also created. If more than one instance of Brigade is
being installed in a given cluster, only the first to be installed must include
these `ClusterRole`s. For the installation of subsequent Brigade instances in
the same cluster, creation of those `ClusterRole`s can be disabled if their
pre-existence proves problematic. To disable the creation of `ClusterRole`s, set
`rbac.installGlobalResources` to `false` at the time of installation. Disabling
this does _not_ disable RBAC, but merely indicates the pre-existence of
cluster-scoped resources.

Brigade itself manages `ServiceAccount`s and other RBAC-related resources for
each Brigade project in each project's own namespace.

## Project Security

Brigade is opinionated about configuring projects and storing data like
credentials. Sensitive information relevant to a project should be set as
[Secrets] on that project. In turn, project secrets are stored in a Kubernetes
secret, so care should be taken in preventing unauthorized access to these
resources.

- Out-of-the-box, credentials should be stored as project secrets
- A project's credentials are accessible to any script running in that project,
  regardless of event.
- For SSH-based Git clones, the SSH key should be stored as a project secret.
  The key for this secret must be `gitSSHKey`.

[Secrets]: /topics/project-developers/secrets

## Script Security

Brigade scripts can indirectly create pods, secrets, and persistent volume
claims. Brigade does not evaluate the security of the containers that a pod
runs. Consequently, it is best to avoid using untrusted containers in Brigade
scripts. Likewise, it is not recommended to inject secrets into a container
without first auditing the container.

## Gateway Security

In Brigade, a gateway is any service that translates some external prompt
(webhook, 3rd party API, cron trigger, etc.) into a Brigade event.

Gateways are the most likely service to have an external network connection. We
suggest the following features of a gateway:

- A gateway should use appropriate network-level encryption
- A gateway should implement authentication/authorization with the upstream service
  wherever appropriate.
  - In most cases, auth requirements should not be passed on to other elements of
    Brigade.
  - The exception is alternative VCS implementations, where the git-sidecar may
    be replaced by another sidecar.

See more info in the [Gateways] doc. There you'll find links to official
Brigade gateways and guidance for writing your own.

[Gateways]: /topics/operators/gateways

