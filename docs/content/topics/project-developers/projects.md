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

In Brigade, a project is a conceptual grouping of event subscriptions paired
with logic expressing how to handle those events.

This document explains how to create and manage projects in Brigade.

## An Introduction to Projects

Brigade projects provide the necessary context for executing Brigade scripts.
In addition to event subscriptions, they also specify the configuration of the
worker in charge of running the event handler logic.

Often times, a Brigade Project will point to an external VCS repository, in
which case this configuration will also be a part of the project definition.
The purpose of this feature is to make it easy to inject source code from a
repository into a Brigade pipeline, and do so in a way that meets standard
expectations about file versioning and storage.

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
apiVersion: brigade.sh/v2
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

There are three high-level sections in the definition above.  They are:

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

```shell
$ brig init --id myproject
```

Or, if the alternate scripting language option of JavaScript is preferred, run:

```shell
$ brig init --id myproject --language js
```

If the project is git-based, supply the git repository name where the Brigade
script for this project will reside:

```shell
$ brig init --id myproject --git https://github.com/<org>/<repo>.git
```

A few assets will be created in the directory in which the command is run.
The bulk of the generated files can be found in the `.brigade/` directory,
including the project definition file (`project.yaml`), a secrets file
(`secrets.yaml`) and a `NOTES.txt` file with next steps.  Additionally, a
`.gitignore` file is created (or amended, if it already exists) to ensure that
the secrets file and script dependencies are not tracked in version control.

### The `brig project create` Command

With a project definition file in hand, you're now ready to create the project
with brig. For purposes of demonstration, let's say the `project.yaml` file
exists in the same directory as the command being run:

```shell
$ brig project create --file project.yaml
```

This command will submit the project definition to Brigade's API server. After
validating that the definition adheres to the project schema, Brigade will
persist the project on the substrate (in the form of a unique namespace) and in
the backing database.

### Update a project

You can update a project at any time with the following command:

```shell
$ brig project update --file project.yaml
```

### Delete a project

To delete a project, run:

```shell
$ brig project delete --id myproject
```

### Listing and inspecting projects with `brig`

You can list all projects via:

```shell
$ brig project list
```

You can also directly inspect your project with `brig project get`. To see the
full project definition, add `--output [yaml|json]`:

```shell
$ brig project get --id myproject --output yaml
```

### Additional project management commands

To manage project secrets, the `brig project secret` suite of commands can be
used. For example, to set secrets for a project via a secrets file, run:

```shell
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
unique namespace a project is assigned, run the `brig project get` command and
note the `kubernetes.namespace` value.  For example:

```plain
$ brig project get --id hello-world --output yaml

apiVersion: brigade.sh/v2
description: The simplest possible example
kind: Project
kubernetes:
  namespace: brigade-97cd352f-90e1-48d0-8797-4f7867a72bd3
metadata:
  created: "2021-08-11T17:47:07.555Z"
  id: hello-world
spec:
  eventSubscriptions:
  - source: brigade.sh/cli
    types:
    - exec
  workerTemplate:
    defaultConfigFiles:
      brigade.js: |
        console.log("Hello, World!");
    logLevel: DEBUG
    useWorkspace: false
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

```shell
$ brig project secrets set --file secrets.yaml
```

### Using other Git providers

Brigade ships with generalized Git support, so use of repositories from any Git
provider should be possible on a Brigade project.

You must ensure, however, that the Kubernetes cluster hosting Brigade can
access the Git repository over the network via the URL provided in `cloneURL`.

To subscribe a Git-based project to events from a corresponding provider, a
[Gateway] is necessary. Brigade currently has gateway support for [GitHub] and
[BitBucket]. See the [Gateways] doc for further info.

[Gateway]: /topics/operators/gateways
[GitHub]: https://github.com/brigadecore/brigade-github-gateway
[BitBucket]: https://github.com/brigadecore/brigade-bitbucket-gateway/tree/v2
[Gateways]: /topics/operators/gateways
