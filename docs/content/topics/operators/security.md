---
title: Securing Brigade
description: 'How to configure security for Brigade.'
aliases:
  - /security.md
  - /intro/security.md
  - /topics/security.md
---

TODO: update per v2

# Securing Brigade

The execution of Brigade scripts involves dynamically creating (and destroying) a
number of Kubernetes objects, including pods, secrets, and persistent volume claims.
For that reason, it is prudent to configure security.

- *Isolate Brigade in a namespace*: It is best to run Brigade in its own namespace. For example,
  in a Helm install, do `helm install --namespace brigade ...`.
- *RBAC Enabled*: When installing with Helm, role-based access control is enabled by default.
  To disable,`--set rbac.enabled=false` will turn off role-based access control.
- *Do not run more than one brigade per namespace*: Running multiple installs of Brigade
  in the same namespace can cause naming collisions which could result in
  unauthorized access to pods.
- *Do not run multiple tenants against the same brigade*: Brigade does not implement
  security controls that allow multiple tenants to share the same Brigade instance.
  Brigade supports multiple projects per Brigade server instance, but those projects
  should be owned by the same tenant.

## How RBAC Is Configured

The Helm [chart](https://github.com/brigadecore/charts/tree/master/charts/brigade) for Brigade
includes an RBAC configuration that is designed to run in an isolated namespace.

The RBAC defines a Service Account for every Deployment (API, Controller, Gateway),
as well as a Service Account for the Brigade worker.

Each deployment also has a Role, which describes the permissions that the particular
service needs.

The resulting Role Bindings, then, are a simple one-to-one match between the
Service Account and the Role.

Workers are an exception. The Service Account for a worker is hard-coded to `brigade-worker`,
and all workers are bound to a Role that allows operations on pods, secrets, and
persistent volume claims.

## Project Security

Brigade is opinionated about configuring projects and storing data like credentials.
Because sensitive information is stored in a project's secret, care should be
taken in preventing unauthorized access to that secret.

- Out-of-the-box, the project is the location where credentials should be stored
- A project's credentials are accessible to any script running in that project,
  regardless of event.
- For SSH-based Git clones, the SSH key should be stored on the project.

Note that if Helm is used to create a project, the project's secrets will be cached
within Helm's release object. Read the [Helm docs](http://helm.sh) to learn how
to secure Helm.

## Script Security

Brigade scripts can create pods, secrets, and persistent volume claims. Brigade does not
evaluate the security of the containers that a pod runs. Consequently, it is best
to avoid using untrusted containers in Brigade scripts. Likewise, it is not recommended
to inject secrets into a container without first auditing the container.

## Gateway Security

In Brigade, a gateway is any service that translates some external prompt (webhook,
3rd party API, cron trigger, etc.) into a Brigade event.

Gateways are the most likely service to have an external network connection. We
suggest the following features of a gateway:

- A gateway should use appropriate network-level encryption
- A gateway should implement authentication/authorization with the upstream service
  wherever appropriate.
  - In most cases, auth requirements should not be passed on to other elements of
    Brigade.
  - The exception is alternative VCS implementations, where the git-sidecar may
    be replaced by another sidecar.

## API Server Security

Brigade includes an API server which allows external tools to discover state about
Brigade projects, builds, and jobs.

This service should only be exposed to the outside network when necessary. And
when exposed, it should use transport layer security (aka SSL) whenever possible.
