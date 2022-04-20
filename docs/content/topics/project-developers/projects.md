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

Brigade projects pair event subscriptions with scripted event-handling logic.
This document explains how to create and manage projects in Brigade.

## Defining a Project

A project definition is represented as YAML or JSON, and may look familiar to
users who have dealt with Kubernetes manifests, although Brigade project
definitions are _not_ Kubernetes manifests.

A typical project definition might resemble this example:

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

Some of the most notable features of this definition include:

  * The `metadata.id` field, which includes a project name that must be unique
    to your instance of Brigade.
  * A `description` of the project.
  * The `spec.eventSubscriptions` configuration, which describes all the events
    the project subscribes to. This particular configuration subscribes to
    events that are created manually from the `brig` CLI.
  * The `spec.workerTemplate` configuration, which describes the container the
    project will launch to handle any events to which it has subscribed. This
    particular configuration includes a `brigade.js` script _in-line_ in the
    `defaultConfigFiles` section and does little more than print "Hello, World!"
    to `stdout` when handling an event originating from the CLI. We'll discuss
    scripting in more detail in the [Scripting Guide].

Writing a script in-line, as in the example above is convenient for very simple
scripts such as this one, but in practical usage can become unwieldy quickly due
to the absence of proper syntax highlighting, for instance, so it is more common
for a `workerTemplate` to reference a git repository where a script can be
found, as in this example:

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
    git:
      cloneURL: https://github.com/example/repo.git
```

> ⚠️&nbsp;&nbsp;Individual Brigade events, from certain sources (such as the
> [GitHub gateway](https://github.com/brigadecore/brigade-github-gateway), for
> instance), can override the project definition's `git.cloneURL` field or
> supplement it by specifying a branch, tag, or commit (by SHA). This ability is
> what makes Brigade capable of implementing CI/CD use cases.

While it is entirely possible to create project definitions from scratch, it
is often more convenient to use `brig init` to generate one for you, which you
can then edit to suit your needs.

For further examples of project definitions to help you get started, see the [Examples](/topics/examples) section of the documentation.

## Creating and Managing Projects

While Brigade project definitions are _not_ Kubernetes manifests and the Brigade
projects they describe are _not_ Kubernetes resources, they can still be thought
of in similar terms. In a Kubernetes cluster, a given resource doesn't exist
until it has been defined by a manifest _and_ that manifest has been applied
(uploaded) to the Kubernetes API server -- and so it is with Brigade projects. A
project is _defined_ in a file and that definition must be uploaded to the
Brigade API server in order to _create_ the project.

With a project definition file in hand -- `project.yaml` in this example -- the
following command will post the new project to the Brigade API server:

```shell
$ brig project create --file project.yaml
```

Projects can be listed:

```shell
$ brig project list
```

A project definition (and status) can be retrieved by specifying the project's
ID and an output format:

```shell
$ brig project get --id <project id> --output yaml
```

Projects can be updated from a modified definition using:

```shell
$ brig project update --file project.yaml
```

To delete a project, run:

```shell
$ brig project delete --id <project id>
```

## Project Namespaces

Brigade creates a unique namespace for each project on the underlying workload
execution substrate (Kubernetes) corresponding to each project. Most Brigade
users will have no need to access this information, but for certain advanced use
cases -- for instance, ones wherein a script is meant to directly modify
Kubernetes resources in a cluster -- a user who has credentials for the
underlying Kubernetes cluster may wish to learn what namespace a project is
assigned to so they can directly modify resources such as Kubernetes
`ServiceAccount`s or `RoleBinding`s in that namespace.

This information can be retrieved from the `kubernetes.namespace` section after
using the `brig project get` command as described in the previous section.

## Project Secrets

The scripts executed by a project's workers often need to make use of sensitive
information that should not be hard-coded into the project definition or script.
To manage such details, the CLI provides a suite of `brig project secret`
commands.

To set a secret:

```shell
$ brig project secret set --project <project id> --set <key>=<value>
```

To set many secrets in bulk from a "flat" JSON or YAML file:

```shell
$ brig project secret set --project <project id> --file secrets.yaml
```

To list secrets for a project:

```shell
$ brig project secret list --project <project id>
```

The above command will display all keys, but will redacted values.

To delete a secret, use:

```shell
$ brig project secret unset --project <project id> --unset <key>
```

The [Scripting Guide] details how to access those secrets within your scripts.

> ⚠️&nbsp;&nbsp;Internally, Brigade never stores secrets in its own database.
> Since Brigade uses Kubernetes to execute scripts and other workloads, it
> stores the secrets as close as possible to where they are used --  namely
> Kubernetes `Secret` resources. This makes sense for a variety of reasons:
>
> 1. If the secrets were stored elsewhere, they'd _still_ need to be copied to
>    Kubernetes `Secret` resources to be usable by the worker `Pod` that
>    executes your script. Storing them there in the first place means they are
>    already where they are needed and there are no additional copies of each
>    secret anywhere else.
>
> 1. Storing secrets _only_ in Kubernetes `Secret` resources means you can trust
>    Brigade with your secrets to the same extent (whatever that may be) that
>    you already trust Kubernetes with your secrets. If you are unhappy with
>    that (for instance, many Kubernetes clusters do not adequately encrypt
>    `Secret` resources) and wish to improve the status quo, you can solve for
>    that at the _cluster_ level. It becomes a Kubernetes problem instead of a
>    Brigade problem.

### Special Secrets for Working with Private Git Repos

Brigade secrets are simple key/value pairs of strings. There are currently
_four_ keys that each offer some special utility in that Brigade itself will
utilize their values if and when required. Specifically, these secrets play a
role in enabling Brigade to clone and work with _private_ git repositories.

* __`gitSSHKey`:__ A PEM-encoded private key beginning with
  `-----BEGIN RSA PRIVATE KEY-----`, ending with
  `-----END RSA PRIVATE KEY-----`, and containing all of its usual line breaks.
  Using a `secrets.yaml` file to set this is the easiest way to preserve all of
  the correct formatting, as in the example below:

  ```yaml
  gitSSHKey: |-
    -----BEGIN RSA PRIVATE KEY-----
    IIEpAIBAAKCAg1wyZD164xNLrANjRrcsbieLwHJ6fKD3LC19E...
    ...
    ...
    -----END RSA PRIVATE KEY-----  
  ```

  If this secret is provided for a given project, Brigade will utilize that key
  when cloning that project's git repository.

  Do _not_ set this secret if you're using a repository URL that begins with
  `https://`.

* __`gitSSHKeyPassword`:__ Optional passphrase for unlocking the PEM-encoded
  private key specified by `gitSSHKey`.

* __`gitUsername`/`gitPassword`:__ Basic auth username and/or password for
  cloning private repos whose URL begins with `https://`.

  If you need to use this, do _not_ set `gitSSHKey`, as `gitSSHKey` takes
  precedence.

  Consult your git provider's documentation for the correct way to set
  username/password for basic auth. For instance, with GitHub, the username is
  ignored/not required and the password should be a
  [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).
  On Bitbucket, by contrast, the username must be a real Bitbucket username (but
  not an email address) and the password should be an
  [app password](https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/).

[Scripting Guide]: /topics/scripting
