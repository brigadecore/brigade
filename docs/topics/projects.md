# Managing Projects in Brigade

In Brigade, a project is just a special Kubernetes secret. The Brigade project
provides a Helm chart that makes it easy to manage projects. This document
explains how to use that chart to manage your Brigade projects.

## An Introduction to Projects

Brigade projects provide the necesary context for executing Brigade scripts.
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

## The Brigade Project Chart

The Brigade Project chart is located in the Brigade source tree at
`brigade-project`. You can also install it out of the Brigade chart repository.

```console
$ helm repo add brigade https://azure.github.io/brigade
$ helm search brigade-project
NAME                   	VERSION	DESCRIPTION
brigade/brigade-project	0.2.0  	Create a Brigade project
```

## Creating and Managing a Project

We recommend using the following pattern to create your project:

### 1. Create a place to store project configs

Store your project configuration in a safe place (probably locally or in a storage
system like Keybase).

```console
$ mkdir -p brigade-projects/myproject
$ cd brigade-projects/myproject
```

> You can store project configs in Git repos, but we don't recommend GitHub for
> this. Keybase has free encrypted private Git repos, which are great for this
> sort of thing.

### 2. Create a `values.yaml` file for your project

```
$ helm inspect values brigade/brigade-project > values.yaml
```

### 3. Edit the values for your project

Read through the generated `values.yaml` file and modify it accordingly.

Our suggestions are as follows, replacing `deis/empty-testbed` with your own project's name:

```yaml
# Definitely do these:
project: "deis/empty-testbed"
secrets: {}

# Probably do these so you can load a GitHub project which has useful stuff
# in it, and if you want to use GitHub webhooks.
repository: "github.com/deis/empty-testbed"
cloneURL: "https://github.com/deis/empty-testbed.git"
github:
   token: "github oauth token"

# If you want GitHub webhooks.
sharedSecret: "IBrakeForSeaBeasts"

# As for the rest, use them if you know you need them.
```

For information on configuring for GitHub, see the [GitHub configuration guide](github.md).

### 4. Install your project

Use Helm to install your chart, with its override values.

```console
$ helm install brigade/brigade-project -n $MY_NAME -f values.yaml
```

Replace `$MY_NAME` with the name of your project (something like `deis-empty-testbed`).

Once the project is created, you can use `brig` or another gateway to begin
writing and running `brigade` scripts.

### 5. Listing and inspect projects with `brig`

If you have the `brig` client installed, you can use it to list and interact with
your projects:

```
$ brig project list
NAME                       	ID                                                            	REPO
technosophos/brigade-trello	brigade-635e505c74ad679bb9144d19950504fbe86b136ac3770bcff51ac6	github.com/technosophos/brigade-trello
deis/empty-testbed         	brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac	github.com/deis/empty-testbed
technosophos/hello-helm    	brigade-b140dc50d4eb9136dccab7225e8fbc9c0f5e17e19aede9d3566c0f	github.com/technosophos/hello-helm
technosophos/twitter-t     	brigade-cf0858d449971e79083aacddc565450b8bf65a2b9f5d66ea76fdb4	github.com/technosophos/twitter-t
```

You can also directly inspect your project with `brig`:

```console
$ brig project get deis/empty-testbed
id: brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
name: deis/empty-testbed
repo:
  name: github.com/deis/empty-testbed
  owner: ""
  cloneurl: https://github.com/deis/empty-testbed.git
  sshkey: ""
kubernetes:
  namespace: default
  vcssidecar: Azure/git-sidecar:latest
  buildStorageSize: "50Mi"
sharedsecret: FakeSharedSecret
github:
  token: 76faketoken789
secrets:
  dbPassword: supersecret
```


### 6. Fetch values later

To get just the values later, you can run `helm get values $MY_NAME`

### 7. Upgrade a project

You can upgrade a project at any time with the command

```console
$ helm upgrade $MY_PROJECT brigade/brigade-project -f values.yaml
```

We suggest not using the `--reuse-values` flag on `helm upgrade` because it can
cause confusing results unless you really know what you are doing.

### 7. Deleting a project

Use `helm delete $MY_PROJECT` to delete a project. Note that once you have done
this, Brigade will no longer execute brigade scripts for this project.

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

In this case, your project's `cloneURL` should be of the form `git@github.com:Azure/brigade.git`
and you will need to add the SSH _private key_ to the `values.yaml` file.

## Using other Git providers

Git providers like BitBucket or GitLab should work fine as Brigade _projects_. However,
the Brigade Gateway does not necessarily support them (yet).

You must ensure, however, that your Kubernetes cluster can access the Git repository
over the network via the URL provided in `cloneURL`.

## Using other VCS systems

It is possible to write a simple VCS sidecar that uses other VCS systems such as
Mercurial, Bazaar, or Subversion. Essentially, a VCS sidecar need only be able
to take the given information from the project and use it to create a local snapshot
of the project in an appointed location. See the [Git sidecar](https://github.com/Azure/brigade/tree/master/git-sidecar)
for an example.
