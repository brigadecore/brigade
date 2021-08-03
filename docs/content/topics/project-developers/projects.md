---
title: Projects
description: How to manage Brigade Projects
section: project-developers
weight: 1
aliases:
  - /projects
  - /topics/projects.md
  - /topics/project-developers/projects.md
---

# Managing Projects in Brigade

In Brigade, a project is a conceptual grouping of event subscriptions paired
with logic expressing how to handle inbound events.

This document explains how to create and manage projects in Brigade.

## An Introduction to Projects

Brigade projects provide the necessary context for executing Brigade scripts.
In addition to event subscriptions, they also specify the configuration of the
worker in charge of running the event handler logic.

Often times, a Brigade Project will point to an external VCS repository, in
which case this configuration will also be a part of the project definition.
The purpose of this feature is to make it easy to inject arbitrary files into
a Brigade pipeline, and do so in a way that meets standard expectations about
file versioning and storage.

Because GitHub is massively popular with open source developers, we chose to
focus on GitHub as a first experience. However, this document explains how to
work with other Git services, and how to extend Brigade to work with other VCS
systems.

## Creating and Managing a Project

Before we can create a Brigade project, we need to define the project itself.
We will discuss the project definition file and then explore how to create and
manage a project from on a given definition file.  If you'd like to skip ahead
and learn how to streamline project creation via the `brig init` command,
proceed to the [brig init section] below.

[brig init section]: #brig-init

### Project definition files

Brigade project definition files are represented in YAML or JSON and follow a
schema that will look familiar to users who have dealt with Kubernetes
manifests - however, they are their own entities and are only understood by
Brigade and the brig CLI.

This approach allows users of Brigade to persist project configuration in a VCS
of their choice.  Updating a pre-existing project is as easy as supplying the
updated configuration to the corresponding brig command.

As an example, let's look at a project definition expressed in YAML and break
it down into its main sections:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/brigadecore/brigade/v2/v2/apiserver/schemas/project.json
apiVersion: brigade.sh/v2-beta
kind: Project
metadata:
  id: hello-world
description: Demonstrates responding to an event with brigadier
spec:
  eventSubscriptions:
  - source: brigade.sh/cli
    types:
    - exec
  workerTemplate:
    logLevel: DEBUG
    defaultConfigFiles:
      brigade.js: |
        const { events } = require("@brigadecore/brigadier");

        events.on("brigade.sh/cli", "exec", async event => {
          console.log("Hello, World!");
        });

        events.process();
```

In addition to the first commented line, which serves to help IDEs provide
auto-completion according to the project schema, there are three high-level
sections in the definition above.  They are:

  1. Project metadata, including:
    i. The `apiVersion` of the schema for this project
    ii. The schema `kind`, which will always be `Project`
    iii. The `id`, or name, of the Project
    iv. A `description` of the project
  2. The `eventSubscriptions` configuration, which contains one event source
    (`brigade.sh/cli`) and one type under that source (`exec`). This particular
    configuration corresponds to events that arise from `brig event create`
    commands.
  3. The `workerTemplate`, which represents the configuration for the worker in
    charge of running the script associated with this project. This particular
    configuration has `logLevel` set to `DEBUG` and the `brigade.js` script for
    this project defined in-line under `defaultConfigFiles`. We'll discuss
    similar scripts in more detail in the [Scripting Guide], but for now we see
    that this script imports the `events` object from the brigadier library and
    declares an event handler for the event source/type combination mentioned
    above.  In response to such events, it prints "Hello, World!" to the
    console.

For further examples of project definition files to help you get started, see
the [examples directory].  Each example sub-directory will have a
`project.yaml` file - this is the project definition file with configuration to
that specific project.

[examples directory]: ./examples
[Scripting Guide]: /topics/scripting

### Brig init

To quickly bootstrap a new project, Brigade offers the `brig init` command. It
will generate a new project definition file based on the options provided,
which can then be used to create the project in Brigade.

For example, to initialize a project named `myproject` with default settings,
which includes TypeScript as the scripting language and no git configuration,
run the following:

```console
$ brig init --id myproject
```

Or, if the alternate scripting language option of JavaScript is preferred, run:

```console
$ brig init --id myproject --language js
```

If the project is git-based, supply the git repository name where the Brigade
script for this project will reside:

```console
$ brig init --id myproject --git https://github.com/<org>/<repo>.git
```

A few assets will be created in the directory in which the command is run.
The bulk of the generated files can be found in the `.brigade` directory,
including the project definition file (`project.yaml`), a secrets file
(`secrets.yaml`) and a `NOTES.txt` file with next steps.  Additionally, a
`.gitignore` file is created to ensure that the secrets file and script
dependencies are not tracked in version control.

### The `brig project create` Command

With a project definition file in hand, you're now ready to create the project
with brig:

```console
$ brig project create --file project.yaml
```

This command will submit the project definition to Brigade's API server. After
validating that the definition adheres to the project schema, Brigade will
persist the project on the substrate (in the form of a unique namespace) and in
the backing database.

### Upgrade a project

You can upgrade a project at any time with the following command:

```console
$ brig project update --file project.yaml
```

### Delete a project

To delete a project, run:

```console
$ brig project delete --id myproject
```

### Listing and inspecting projects with `brig`

You can list all projects via:

```console
$ brig project list
```

You can also directly inspect your project with `brig project get`:

```console
$ brig project get --id myproject
```

### Additional project management commands

To manage project secrets, the `brig project secret` suite of commands can be
used. For example, to set secrets for a project via a secrets file, run:

```console
$ brig project secret set --file secrets.yaml
```

To explore different ways to manage secrets in Brigade, see the [Secrets] doc.

To manage roles for a project, see the `brig project roles` suite of commands.
Roles can be created, listed and revoked. For an overview on roles and
authorization in general, see the [Authorization] doc.

[Secrets]: /topics/project-developers/secrets
[Authorization]: /topics/administrators/authorization

## Project namespaces

Brigade creates a unique namespace on the underlying substrate (Kubernetes)
corresponding to each project. Although most users shouldn't have a need to
inspect resources under a project namespace on the substrate, to see which
unique namespace corresponds to which project, run the following `kubectl`
command:

```console
$ kubectl get namespaces --show-labels
NAME                                           STATUS   AGE     LABELS
brigade-9f88dd8d-b804-4558-a26f-5e9ec32f1628   Active   8m13s   brigade.sh/project=myproject
```

## Git-based projects

### Using SSH Keys

You can use SSH keys and a `git+ssh` URL to secure a private repository.

In this case, your project's `cloneURL` should be of the form
`git@github.com:<org>/<repo>.git` and you will need to add the SSH
_private key_ as a secret to the project with the key `gitSSHKey`.

For example, if project secrets are contained in a `secrets.yaml` file, the
private key would be added like so:

```yaml
gitSSHKey: |-
  -----BEGIN RSA PRIVATE KEY-----
  IIEpAIBAAKCAg1wyZD164xNLrANjRrcsbieLwHJ6fKD3LC19E...
  ...
  ...
  -----END RSA PRIVATE KEY-----
```

The project secrets can then be updated via the usual brig command:

```console
$ brig project secrets set --file secrets.yaml
```

### Using other Git providers

Brigade currently has gateway support for [GitHub] and [BitBucket]. Other
providers should work fine for Brigade projects, though gateways to handle
webhooks might not yet exist. See the [Gateways] doc for further info.

You must ensure, however, that the Kubernetes cluster hosting Brigade can
access the Git repository over the network via the URL provided in `cloneURL`.

[GitHub]: https://github.com/brigadecore/brigade-github-gateway
[BitBucket]: https://github.com/brigadecore/brigade-bitbucket-gateway/tree/v2
[Gateways]: /topics/operators/gateways

### Using other VCS systems

It is possible to write a VCS sidecar that uses other VCS systems such as
Mercurial, Bazaar, or Subversion. Essentially, a VCS sidecar need only be able
to take the given information from the project and use it to create a local
snapshot of the project in an appointed location. See the default
[Git initializer] code for an example.

[Git initializer]: https://github.com/brigadecore/brigade/tree/v2/v2/git-initializer