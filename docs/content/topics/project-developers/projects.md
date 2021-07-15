---
title: Projects
description: How to manage Brigade Projects
aliases:
  - /projects.md
  - /topics/projects.md
  - /topics/project-developers/projects.md
---

TODO: update per v2

# Managing Projects in Brigade

In Brigade, a project is just a special Kubernetes secret. The Brigade project currently offers
two methods to create a project: via the `brig` cli and via the [brigade-project][brigade-project-chart]
Helm chart.  The latter is managed in the [brigadecore/charts][charts] repo
and an in-depth overview of its configuration can be seen in the chart
[README](https://github.com/brigadecore/charts/blob/master/chart/brigade-project/README.md).

This document explains how to use both methods for managing your Brigade projects.

## An Introduction to Projects

Brigade projects provide the necessary context for executing Brigade scripts.
They provide _permission_ to run scripts, _authentication_ for some operations,
_configuration_ for VCS, and _secret management_ for Brigade scripts.

Often times, a Brigade Project will point to an external VCS repository. The
purpose of this feature is to make it easy to inject arbitrary files into a
Brigade pipeline, and do so in a way that meets standard expectations about
file versioning and storage.

Because GitHub is massively popular with open source developers, we chose to focus
on GitHub as a first experience. However, this document explains how to work with
other Git services, and how to extend Brigade to work with other VCS systems.

## A Few Recommendations

- Name your project with the GitHub convention `userOrGroup/project`.
- When installing the Helm chart, use the `-n NAME` flag to name your Helm release.
  We suggest `userOrGroup-project` as the release name just to make it easy.
- When you use a GitHub repo, don't put secrets in it, even if you make your
  repo private. Put them in your `values.yaml` file for your project (and don't
  put that in GitHub).

## Creating and Managing a Project

### The `brig project create` Command

The preferred default for project creation and maintenance is via `brig`.  To create a
new Brigade project, simply run `brig project create` and you will be prompted to answer a few questions:

```console
$ brig project create
? VCS or no-VCS project? VCS
? Project name brigadecore/empty-testbed
? Full repository name github.com/brigadecore/empty-testbed
? Clone URL (https://github.com/your/repo.git) https://github.com/brigadecore/empty-testbed.git
? Add secrets? No
Auto-generated a Shared Secret: "FweBxcwJvcbTTuW5CquyPtHM"
? Configure GitHub Access? No
? Configure advanced options No
```

You can use `--dry-run --verbose` to see the answers to the question without creating
a new release. For more, see `brig project create --help`.

You can optionally customize a bunch of advanced options during `brig project create`:

- *Custom VCS sidecar*: The default sidecar uses Git to fetch your repository
- *Build storage size*: By default, 50Mi of shared temp space is allocated per build. Larger values slow down build startup. Units are Ki, Mi, or Gi
- *Build storage class*: Kubernetes provides named storage classes. If you want to use a custom storage class for Builds, set the class name here
- *Job cache storage class*: Same as before, the custom storage class that will be used for job caches
- *SecretKeyRef usage*: Allow or disallow usage of secretKeyRef in job environments
- *Worker image registry or DockerHub org*: For non-DockerHub, this is the root URL. For DockerHub, it is the org
- *Worker image name*: The name of the worker image, e.g. workerImage
- *Custom worker image tag*: The worker image tag to pull, e.g. 1.2.3 or latest
- *Worker image pull policy*: The image pull policy determines how often Kubernetes will try to refresh this image
- *Worker command*: Override the worker's default command (yarn -s start)
- *Initialize Git submodules*: For repos that have submodules, initialize them on each clone. Not recommended on public repos
- *Allow host mounts*: Allow host-mounted volumes for worker and jobs. Not recommended in multi-tenant clusters
- *Allow privileged jobs*: Allow jobs to mount the Docker socket or perform other privileged operations. Not recommended for multi-tenant clusters
- *Image pull secrets*: Comma-separated list of image pull secret names that will be supplied to workers and jobs
- *Default script ConfigMap name*: It is possible to store a default script in a ConfigMap. Supply the name of that ConfigMap to use the script
- *brigade.js file path relative to the repository root*: brigade.js file path relative to the repository root, e.g. 'mypath/brigade.js'. Absolute paths will not be accepted
- *Upload a default brigade.js script*: The local path to a default brigade.js file that will be run if none exists in the repo. Overrides the ConfigMap script
- *Default config ConfigMap name*: It is possible to store a default brigade.json config in a ConfigMap. Supply the name of that ConfigMap to use the config
- *brigade.json file path relative to the repository root*: brigade.json file path relative to the repository root, e.g. 'mypath/brigade.json'. Absolute paths will not be accepted
- *Upload a default brigade.json config*: The local path to a default brigade.json config file that will be used if none exists in the repo. Overrides the ConfigMap config

#### Storing a Script in a ConfigMap

It is possible to store a `brigade.js` script in a dedicated ConfigMap, and then share that ConfigMap with multiple projects.

For example, if you have a local file named `examples/brigade.js`, you can use the `kubectl` command to turn it into a ConfigMap for Brigade to consume:

```console
$ kubectl create configmap default-brigade-script --from-file=brigade.js=examples/brigade.js
```

The command above stores the `examples/brigade.js` script in a ConfigMap named `default-brigade-script`. When creating a new project, you can answer the *Default script ConfigMap name* question by providing the name of this ConfigMap: `default-brigade-script`.

If this value is supplied, the `brigade.js` in the ConfigMap will be used if and only if neither the project nor the repository provides an alternative `brigade.js` file. In other words, this is the _last location_ checked for a Brigade script.

This same pattern can be used to store a default `brigade.json` configuration file in a ConfigMap for use by multiple projects. The associated prompt is *Default brigade.json config ConfigMap name*

#### Storing a Default Script in the Project

It is also possible to store a default Brigade script within the project definition. When `brig project create` prompts for *Upload a default brigade.js script* you can enter the local path to a `brigade.js` file (e.g. `examples/brigade.js`) and that file will be copied into the project. Note that if you edit the `brigade.js` you will need to edit the project's Kubernetes Secret and update the script accordingly.

If a script is set here, it will override a ConfigMap based script. However, if a repository holds a `brigade.js` file, that will take precedence over this.

This same pattern can be used to store a default `brigade.json` configuration file for a project.  The associated prompt is *Upload a default brigade.json config*.

### Managing Your Projects

With `brig project create`, your projects are stored in Kubernetes secrets. You
can save a local copy of that secret using `brigade project create --out=myproject.json`.

If you have already created the secret, you can fetch it from Kubernetes by running
`brig project get my/project` where `my/project` is the project name you assigned.

## Creating and Managing a Project (The Old Way)

Note: Managing Brigade projects via Helm chart is being deprecated in favor of using `brig`.

### The Brigade Project Chart

The Brigade Project chart is located in the [brigadecore/charts][charts] source tree at
`charts/brigade-project`. You can also install it out of the Brigade chart repository.

```console
$ helm repo add brigade https://brigadecore.github.io/charts
$ helm search brigade/brigade-project
NAME                   	CHART VERSION	APP VERSION	DESCRIPTION
brigade/brigade-project	1.0.0       	v1.0.0    	Create a Brigade project
```

We recommend using the following pattern to create your project:

#### 1. Create a place to store project configs

Store your project configuration in a safe place (probably locally or in a storage
system like Keybase).

```console
$ mkdir -p brigade-projects/myproject
$ cd brigade-projects/myproject
```

> You can store project configs in Git repos, but we don't recommend GitHub for
> this. Keybase has [free encrypted private Git repos](https://keybase.io/blog/encrypted-git-for-everyone),
> which are great for this sort of thing.

#### 2. Create a `values.yaml` file for your project

```
$ helm inspect values brigade/brigade-project > values.yaml
```

#### 3. Edit the values for your project

Read through the generated `values.yaml` file and modify it accordingly.

Our suggestions are as follows, replacing `brigadecore/empty-testbed` with your own project's name:

```yaml
# Definitely do these:
project: "brigadecore/empty-testbed"
secrets: {}

# Probably do these so you can load a GitHub project which has useful stuff
# in it, and if you want to use GitHub webhooks.
repository: "github.com/brigadecore/empty-testbed"
cloneURL: "https://github.com/brigadecore/empty-testbed.git"
github:
   token: "github oauth token"

# If you want GitHub webhooks.
sharedSecret: "IBrakeForSeaBeasts"

# As for the rest, use them if you know you need them.
```

For information on configuring for GitHub, see the [GitHub configuration guide](github.md).

#### 4. Install your project

Use Helm to install your chart, with its override values.

```console
$ helm install brigade/brigade-project -n $MY_NAME -f values.yaml
```

Replace `$MY_NAME` with the name of your project (something like `brigadecore-empty-testbed`).

Once the project is created, you can use `brig` or another gateway to begin
writing and running `brigade` scripts.

#### 5. Fetch values later

To get just the values later, you can run `helm get values $MY_NAME`

#### 6. Upgrade a project

You can upgrade a project at any time with the command

```console
$ helm upgrade $MY_PROJECT brigade/brigade-project -f values.yaml
```

We suggest not using the `--reuse-values` flag on `helm upgrade` because it can
cause confusing results unless you really know what you are doing.

#### 7. Deleting a project

Use `helm delete $MY_PROJECT` to delete a project. Note that once you have done
this, Brigade will no longer execute brigade scripts for this project.


## Listing and inspecting projects with `brig`

If you have the `brig` client installed, you can use it to list and interact with
your projects:

```
$ brig project list
NAME                       	ID                                                            	REPO
technosophos/brigade-trello	brigade-635e505c74ad679bb9144d19950504fbe86b136ac3770bcff51ac6	github.com/technosophos/brigade-trello
brigadecore/empty-testbed         	brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac	github.com/brigadecore/empty-testbed
technosophos/hello-helm    	brigade-b140dc50d4eb9136dccab7225e8fbc9c0f5e17e19aede9d3566c0f	github.com/technosophos/hello-helm
technosophos/twitter-t     	brigade-cf0858d449971e79083aacddc565450b8bf65a2b9f5d66ea76fdb4	github.com/technosophos/twitter-t
```

You can also directly inspect your project with `brig`:

```console
$ brig project get brigadecore/empty-testbed
id: brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
name: brigadecore/empty-testbed
repo:
  name: github.com/brigadecore/empty-testbed
  owner: ""
  cloneurl: https://github.com/brigadecore/empty-testbed.git
  sshkey: ""
kubernetes:
  namespace: default
  vcssidecar: brigadecore/git-sidecar:latest
  buildStorageSize: "50Mi"
sharedsecret: FakeSharedSecret
github:
  token: 76faketoken789
secrets:
  dbPassword: supersecret
```


## Internal Brigade Project Names

Brigade creates an "internal name" for each project. It looks something like
this:

```
brigade-635e505c74ad679bb9144d19950504fbe86b136ac3770bcff51ac6
```

There is nothing fancy about this name. In fact, it is just a prefixed hash of
the project name. Its purpose is merely to ensure that we can meet the naming
requirements of Kubernetes without imposing undue restrictions on how you name
your project.

_These names are intentionally repeatable._ Two projects with the same name should
also have the same internal name.

## Using SSH Keys

You can use SSH keys and a `git+ssh` URL to secure a private repository.

In this case, your project's `cloneURL` should be of the form `git@github.com:brigadecore/brigade.git`
and you will need to add the SSH _private key_ to the `values.yaml` file.

When doing `brig project create`, URLs that do not use HTTP or HTTPS will prompt
for (optionally) adding an SSH key.

## Using other Git providers

Git providers like BitBucket or GitLab should work fine as Brigade _projects_. However,
the [Brigade GitHub Gateway](./github.md) does not necessarily support them (yet).

You must ensure, however, that your Kubernetes cluster can access the Git repository
over the network via the URL provided in `cloneURL`.

## Using other VCS systems

It is possible to write a simple VCS sidecar that uses other VCS systems such as
Mercurial, Bazaar, or Subversion. Essentially, a VCS sidecar need only be able
to take the given information from the project and use it to create a local snapshot
of the project in an appointed location. See the [Git sidecar](https://github.com/brigadecore/brigade/tree/master/git-sidecar)
for an example.

[charts]: https://github.com/brigadecore/charts
[brigade-project-chart]: https://github.com/brigadecore/charts/tree/master/charts/brigade-project