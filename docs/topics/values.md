# Project Values

When creating a new Brigade project, you will be asked to configure the project's
chart. This guide covers the configuration directives.

## Finding Configuration Options

Brigade projects can be easily installed via Helm. And Helm provides a useful
tool for generating a stub configuration file, too:

```console
$ helm inspect values brigade/brigade-project > values.yaml
```

After running the command above, you may edit the `values.yaml` file and then
later install the chart with:

```console
$ helm install -n my-project brigade/brigade-project -f values.yaml
```

See the [Projects Guide](projects.md) for more.

## The Configurable Parameters

This section explains each of the configuration directives available in the
`values.yaml` for a Brigade project.

A quick note on `deis/empty-testbed`: Early in our development cycle, we created
a simple GitHub repository that we could use for testing. Over time, that repository
made its way into our documentation and demos. You are free to use it for testing.

### `project` (REQUIRED)

The `project` parameter is used to give your project a human-readable name. The
convention for the project name is `user/project` or `org/project`, as you would
see with GitHub, Dockerhub, and similar sites.

```yaml
project: "deis/empty-testbed"
```

### `repository` (RECOMMENDED)

One of the features of Brigade is that it can use source code repositories to
store auxiliary files. For CI/CD-like systems, your source code may go here.
For other Brigade pipelines, such a directory may merely hold supporting resources.

You are not required to have a repository set up for every project. But we do
recommend this as a best practice for attaching auxiliary files in a version-controlled
way.

While Brigade has a pluggable VCS system, it only ships with GitHub (and Git)
support.

The `repository` directive is a protocol-neutral repository name. For
GitHub projects, it should always been in the form `github.com/ORG/PROJECT`.

```yaml
repository: "github.com/deis/empty-testbed"
```

This values is used by gateways to construct API calls to upstream VCS API
services (e.g. the GitHub API). Any time you are using GitHub, you _should_
provide this value.

This field is case-sensitive.

### `cloneURL` (RECOMMENDED)

As discussed in the section above, Brigade can use a VCS to store auxiliary
files, such as source code.

Set the `cloneURL` to the location where you want Brigade to fetch a copy of
your source code repository.

```yaml
cloneURL: "https://github.com/deis/empty-testbed.git"
```

Note that there is no firm requirement that `cloneURL` and `repository` point to
the same domain. Typically they do, though.

A `cloneURL` can support any form of URL that the upstream provided supports. So,
for GitHub, this can be a `https://...` URL or a `git@github.com...` URL.

The `cloneURL` setting is used by the VCS initialization container to fetch
a copy of the repository for each job in the build. For example, the default
VCS sidecar, `git-sidecar`, will use the `cloneURL` to fetch a shallow clone
of the git repository.

### `initGitSubmodules` (OPTIONAL)

Determine if git will initialize all submodules in the repository. Default: false

### `sharedSecret` (OPTIONAL)


This value is used by GitHub and other services to compute hook HMACs. This
is one way of preventing data tampering during transmission.

```yaml
sharedSecret: "IBrakeForSeaBeasts"
```

In the future, this may be moved into gateway-specific settings.

### The `github` section (OPTIONAL)

This section controls the GitHub gateway configuration.

#### `token` (OPTIONAL)

If you are using the GitHub gateway, the `token` directive has the OAuth2 token
used to authenticate Brigade to GitHub. Certain upstream requests (like notifications
and fetching the `brigade.js`) use the token.

```yaml
github:
   token: "github oauth token"
```

### `sshKey` (OPTIONAL)

The `sshKey` is used to provide an SSH key that is paired with the `cloneURL` to
fetch repository information over a secured SSH connection.

```yaml
 sshKey: |-
  -----BEGIN RSA PRIVATE KEY-----
  IIEpAIBAAKCAg1wyZD164xNLrANjRrcsbieLwHJ6fKD3LC19E...
  ...
  ...
  -----END RSA PRIVATE KEY-----
```

This is used by a VCS sidecar.

### The `secrets` section

Brigade provides a way for you to pass _ad hoc_ name/value pairs from your
`values.yaml` file to the `brigade.js`. Passwords, tokens, and sensitive information
_ought_ to be handled this way. But you may also choose to place other data here.

```yaml
secrets:
  myPassword: superSecret
```

Within your `brigade.js` scripts, the `secrets` data is accessible as a property
on the `Project` object:

```javascript
events.on("exec", (e, p) => {
  var job = new Job("j", "alpine:3.7")
  job.env = {
    MY_PASSWORD: p.secrets.myPassword
  }
})
```

This will mount the secret onto your Job, and expose it as an environment variable
named `$MY_PASSWORD`.

### `namespace` (OPTIONAL)

This controls the namespace into which your builds will be deployed. This is
considered an expert option, and you may need to manually adjust RBACs if you
set this.

```yaml
namespace: "default"
```

### `vcsSidecar` (OPTIONAL)

This allows you to replace the default `git-sidecar` with your own custom VCS
sidecar. A sidecar image is given access to a few specific variables, including
`cloneURL` and `sshKey`. It is expected to make the data at the `cloneURL`'s source
available locally.

### `buildStorageSize` (OPTIONAL)

This allows one to set the size of the build shared storage space used by the jobs.

### `allowPrivilegedJobs` (REQUIRED)

Determine whether the jobs in this project are allowed to go into privileged mode.
Privileged mode is determined by the underlying cluster. Kubernetes, for example,
allows a privileged pod to mount the Docker socket or run Docker-in-Docker.

```yaml
allowPrivilegedJobs: "true"
```

If this is set to `true`, then within your `brigade.js` script, you still must also
turn on privileged mode for each `Job` that needs it. Privileged mode is never
enabled by default.

If this is set to `false`, no Job will be allowed to go into privileged mode even
of the `brigade.js` file sets the appropriate flag.

### `allowHostMounts` (OPTIONAL)

If this is `true` and `allowPrivilegedJobs` is true, then script authors may
not only turn on privileged mode, but may also get a copy of the Docker socket
mounted to the pod. This is potentially dangerous to the cluster's Docker stability
if misused, so it is `false` by default.

We recommend using Docker-in-Docker instead when possible.

