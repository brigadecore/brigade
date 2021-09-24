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
- *RBAC Enabled*: Role-based access control is enabled by default. While it is
  possible to disable via a corresponding chart value, we won't vouch for the
  security of anything Brigade-related in a non-RBAC-enabled cluster.
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

[Helm chart]: https://github.com/brigadecore/brigade/tree/v2/charts/brigade
[ingress controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
[Deployment]: /topics/operators/deploy

## How RBAC Is Configured

The [Helm chart] for Brigade includes all RBAC configuration necessary to run
properly. The RBAC resources include a service account for every Brigade core
deployment (API, Observer, Scheduler, etc.), as well as a service accounts for
each project: one for the Brigade worker and another for the jobs. This allows
competent Kubernetes users with adequate permissions to modify the service
account(s) for a given project to permit things that cannot be done otherwise--
such as scripting Kubernetes itself.

Each aforementioned deployment also has a Role, which describes the permissions
that the particular service needs. The resulting Role Binding resources, then,
are a one-to-one match between the Service Account and the Role.

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

