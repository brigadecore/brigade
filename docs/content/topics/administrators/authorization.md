---
title: Authorization
description: Authorization setup for Brigade
section: administrators
weight: 2
aliases:
  - /authorization
  - /topics/authorization.md
  - /topics/administrators/authorization.md
---

# Brigade Authorization

Authorization in Brigade consists of roles with particular scopes, which are
granted to users and service accounts. When users interact with Brigade via
the brig CLI or when a service account interacts with Brigade via an SDK,
Brigade checks to be sure the requestor is sufficiently authorized before
proceeding.

The three core authorization components in Brigade are:

  * [Users](#users)
  * [Service Accounts](#service-accounts)
  * [Roles](#roles)

Users are generated in the system after successful authentication with the
selected third-party auth provider and the creation of service accounts and
role assignments is the responsibility of the administrator of Brigade.

> Note: There is one method to have role assignments be automatically granted
for a given set of users upon their first login to Brigade. The role
specifically grants system-level admin privileges to each designated user.
Details on configuring this deployment option can be seen in the
[Authentication] doc.

[Authentication]: /topics/administrators/authentication

## Users

A User in Brigade represents a human user authenticated into the system via the
third-party auth provider selected during Brigade's deployment. There is no
mechanism to create users outside of this authentication system. Users are
assigned [roles](#roles) granting scoped permissions around their interactions
with resources in Brigade.

In addition to viewing a particular user's details, administrators can list,
lock, unlock and delete users. All of these management functions exist under
the `brig users` suite of commands. To see the full suite, issue the following
help command:

```console
$ brig users --help
```

## Service Accounts

A Service Account in Brigade represents a non-human actor that can be assigned
a [role](#roles) granting scoped permissions for interacting with resources in
Brigade. A common pattern is to create a service account for a gateway and
assign it an `EVENT_CREATOR` role such that it may submit events into Brigade.

Administrators may create, list, get, lock, unlock and delete service accounts.
All of these management functions exist under the `brig service-accounts` suite
of commands. To see the full suite, issue the following help command:

```console
$ brig service-accounts --help
```

## Roles

A Role in Brigade represents a scoped set of permissions around resource access
within Brigade, which can then be assigned to a [User](#users) or
[Service Account](#service-accounts). There exist system-level roles as well as
project-level roles.

Administrators may grant, revoke and list roles, either at the system-level or
the project-level. All of these management functions exist under the
`brig roles` or `brig project roles` suites of commands. To see the full
suites, issue the following help commands:

```console
$ brig roles --help
$ brig project roles --help
```

### System-level Roles

System-level roles in Brigade are as follows:

  * `ADMIN` - Enables system management including system-level permissions for
    other users and service accounts.
  * `EVENT_CREATOR`- Enables creation of events for all projects. An event
    `source` must be provided for each assignment of this role.
  * `PROJECT_CREATOR` - Enables creation of new projects. When a user with this
    role assignment creates a new project, they are automatically granted all
    [project-level roles](#project-level-roles) for that project.
  * `READER`- Enables global read-only access to Brigade.

Each role is itself a sub-command under `brig role grant` or
`brig role revoke`. For example, to grant the `ADMIN` role to user `Mary`, the
following command would be issued:

```console
$ brig role grant ADMIN --user Mary
```

Any system role may also be granted to a service account.

### Project-level Roles

Project-level roles in Brigade are as follows:

  * `PROJECT_ADMIN` - Enables management of all aspects of the project,
    including its secrets, as well as project-level permissions for other users
    and service accounts.
  * `PROJECT_DEVELOPER` - Enables updating the project definition, but does NOT
    enable management of the project's secrets or project-level permissions for
    other users and service accounts.
  * `PROJECT_USER` - Enables creation and management of events associated with
    the project

Each role is itself a sub-command under `brig project role grant` or
`brig project role revoke`. For example, to grant the `PROJECT_ADMIN` role to
user `Mary` for project `Arecibo`, the following command would be issued:

```console
$ brig project role grant ADMIN --id Arecibo --user Mary
```

Any project role may also be granted to a service account.