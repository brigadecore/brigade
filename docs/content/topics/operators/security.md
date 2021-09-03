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

The execution of Brigade scripts involves dynamically creating (and destroying) a
number of Kubernetes objects, including pods, secrets, and persistent volume claims.
For that reason, it is prudent to configure security.

- *Isolate Brigade in a namespace*: It is best to run Brigade in its own namespace. For example,
  in a Helm install, do `helm install --namespace brigade ...`.
- *RBAC Enabled*: When installing with Helm, role-based access control is enabled by default.
  To disable,`--set rbac.enabled=false` will turn off role-based access control.
- *Do not run more than one Brigade per namespace*: Running multiple installs of Brigade
  in the same namespace can cause naming collisions which could result in
  unauthorized access to pods.

## API Server Security

If Brigade's API server will be exposed to the internet, e.g. via an external
IP (recall that Brigade's default service type in the [Brigade chart] is
`LoadBalancer` and so a public-facing IP will be provisioned), care should be
taken to secure inbound communication. Minimally, TLS should be enabled and 
SSL certificates should not be self-signed. You may also consider using an
[Ingress Controller] for an extra layer of traffic filtering and/or protection.

For more details on deploying Brigade securely, see the [Deployment] doc.

[Brigade chart]: https://github.com/brigadecore/brigade/tree/v2/charts/brigade
[Ingress Controller]: https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/
[Deployment]: /topics/operators/deploy

## How RBAC Is Configured

The Helm [chart][Brigade chart] for Brigade includes an RBAC configuration that
is designed to run in an isolated namespace.

The RBAC defines a Service Account for every Deployment (API, Observer,
Scheduler, etc.), as well as a Service Account for the Brigade worker.

Each deployment also has a Role, which describes the permissions that the
particular service needs.

The resulting Role Bindings, then, are a simple one-to-one match between the
Service Account and the Role.

Workers are an exception. The Service Account for a worker is hard-coded to
`workers` and all workers are bound to a Role that allows operations on pods,
secrets, and persistent volume claims.

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

[Secrets]: /topics/project-developers/secrets

## Script Security

Brigade scripts can create pods, secrets, and persistent volume claims. Brigade
does not evaluate the security of the containers that a pod runs. Consequently,
it is best to avoid using untrusted containers in Brigade scripts. Likewise, it
is not recommended to inject secrets into a container without first auditing
the container.

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

See more info in the [Gateways] doc. Their you'll find links to official
Brigade gateways and guidance for writing your own.

[Gateways]: /topics/operators/gateways

