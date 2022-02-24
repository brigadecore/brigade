---
linkTitle: Quickstart
title: A Brigade Quickstart
description: A Brigade Quickstart.
section: intro
weight: 3
aliases:
  - /quickstart.md
  - /intro/quickstart.md
---

This QuickStart presents a comprehensive introduction to Brigade. You will
install Brigade with default configuration on a local, development-grade
cluster, create a project and an event, watch Brigade handle that event, then
clean up.

If you prefer learning through video, check out the
[video adaptation](https://www.youtube.com/watch?v=VFyvYOjm6zc) of this guide on
our YouTube channel.

* [Prerequisites](#prerequisites)
* [Install Brigade](#install-brigade)
  * [Install the Brigade CLI](#install-the-brigade-cli)
  * [Install Server-Side Components](#install-server-side-components)
  * [Port Forwarding](#port-forwarding)
* [Trying It Out](#trying-it-out)
  * [Log into Brigade](#log-into-brigade)
  * [Create a Project](#create-a-project)
  * [Create an Event](#create-an-event)
* [Cleanup](#cleanup)
* [Next Steps](#next-steps)
* [Troubleshooting](#troubleshooting)

## Prerequisites

> ⚠️ The default configuration used in this guide is appropriate only for
> evaluating Brigade on a local, development-grade cluster and is not
> appropriate for _any_ shared cluster -- _especially a production one_. See our
> [Deployment Guide](/topics/operators/deploy/) for instructions suitable for
> shared or production clusters. We have tested these instructions on a local
> [KinD](https://kind.sigs.k8s.io/) cluster.

* A local, _development-grade_ Kubernetes v1.16.0+ cluster
* [Helm v3.7.0+](https://helm.sh/docs/intro/install/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
* Free disk space. The installation requires sufficient free disk space and will
  fail if your disk is nearly full.

## Install Brigade

### Install the Brigade CLI

In general, the Brigade CLI, `brig`, can be installed by downloading the
appropriate pre-built binary from our
[releases page](https://github.com/brigadecore/brigade/releases) to a directory
on your machine that is included in your `PATH` environment variable. On some
systems, it is even easier than this. Below are instructions for common
environments:

**Linux**

```bash
$ curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-linux-amd64
$ chmod +x /usr/local/bin/brig
```

**macOS**

The popular [Homebrew](https://brew.sh/) package manager provides the most
convenient method of installing the Brigade CLI on a Mac:

```bash
$ brew install brigade-cli
```

Alternatively, you can install manually by directly downloading a pre-built
binary:

```bash
$ curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-darwin-amd64
$ chmod +x /usr/local/bin/brig
```

**Windows**

```powershell
> mkdir -force $env:USERPROFILE\bin
> (New-Object Net.WebClient).DownloadFile("https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-windows-amd64.exe", "$ENV:USERPROFILE\bin\brig.exe")
> $env:PATH+=";$env:USERPROFILE\bin"
```

The script above downloads `brig.exe` and adds it to your `PATH` for the current
session. Add the following line to your
[PowerShell Profile](https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/)
if you want to make the change permanent:

```powershell
> $env:PATH+=";$env:USERPROFILE\bin"
```

### Install Server-Side Components

To install server-side components on your local, development-grade cluster:

1. Enable Helm's experimental OCI support:

    **POSIX**
    ```bash
    $ export HELM_EXPERIMENTAL_OCI=1
    ```

    **PowerShell**
    ```powershell
    > $env:HELM_EXPERIMENTAL_OCI=1
    ```

1. Run the following commands to install Brigade with default configuration:

    ```
    $ helm install brigade \
        oci://ghcr.io/brigadecore/brigade \
        --version v2.3.1 \
        --create-namespace \
        --namespace brigade \
        --wait \
        --timeout 300s
    ```

    > ⚠️ Installation and initial startup may take a few minutes to complete.

    If the deployment fails, proceed to the [troubleshooting](#troubleshooting)
    section.

### Port Forwarding

Since you are running Brigade locally, use port forwarding to make the Brigade
API available via the local network interface:

**POSIX**
```bash
$ kubectl --namespace brigade port-forward service/brigade-apiserver 8443:443 &>/dev/null &
```

**PowerShell**
```powershell
> kubectl --namespace brigade port-forward service/brigade-apiserver 8443:443 *> $null  
```

## Trying It Out

### Log into Brigade

To authenticate to Brigade as the root user, you first need to acquire the
auto-generated root user password:

**POSIX**
```bash
$ export APISERVER_ROOT_PASSWORD=$(kubectl get secret --namespace brigade brigade-apiserver --output jsonpath='{.data.root-user-password}' | base64 --decode)
```

**PowerShell**
```powershell
> $env:APISERVER_ROOT_PASSWORD=$(kubectl get secret --namespace brigade brigade-apiserver --output jsonpath='{.data.root-user-password}' | base64 --decode)
```

Then:

**POSIX**
```bash
$ brig login --insecure --server https://localhost:8443 --root --password "${APISERVER_ROOT_PASSWORD}"
```

**PowerShell**
```powershell
> brig login --insecure --server https://localhost:8443 --root --password "$env:APISERVER_ROOT_PASSWORD"
```

The `--insecure` flag instructs `brig login` to ignore the self-signed
certificate used by our local installation of Brigade.

If the `brig login` command hangs or fails, double-check that port-forwarding
for the `brigade-apiserver` service was successfully completed in the previous
section.

### Create a Project

A Brigade [project](/topics/project-developers/projects) pairs event
subscriptions with worker (event handler) configuration.

1. Rather than create a project definition from scratch, we'll accelerate the
   process using the `brig init` command:

    ```console
    $ mkdir first-project
    $ cd first-project
    $ brig init --id first-project
    ```

    This will create a project definition similar to the following in
    `.brigade/project.yaml`. It subscribes to `exec` events emitted from a
    source named `brigade.sh/cli`. (This type of event is easily created using
    the CLI, so it is great for demo purposes.) When such an event is received,
    the embedded script is executed. The script itself branches depending on the
    source and type of the event received. For an `exec` event from the source
    named `brigade.sh/cli`, this script will spawn and execute a simple "Hello
    World!" job. For any other type of event, this script will do nothing.

    ```yaml
    apiVersion: brigade.sh/v2
    kind: Project
    metadata:
      id: first-project
    description: My new Brigade project
    spec:
      eventSubscriptions:
        - source: brigade.sh/cli
          types:
            - exec
    workerTemplate:
      logLevel: DEBUG
      defaultConfigFiles:
      brigade.ts: |
        import { events, Job } from "@brigadecore/brigadier"
        
        // Use events.on() to define how your script responds to different events. 
        // The example below depicts handling of "exec" events originating from
        // the Brigade CLI.
        
        events.on("brigade.sh/cli", "exec", async event => {
            let job = new Job("hello", "debian:latest", event)
            job.primaryContainer.command = ["echo"]
            job.primaryContainer.arguments = ["Hello, World!"]
            await job.run()
        })

        events.process()
    ```

1. The previous command only generated a project definition from a template. We
   still need to upload this definition to Brigade to complete project creation:

    ```console
    $ brig project create --file .brigade/project.yaml
    ```

1. To see that Brigade now knows about this project, use `brig project list`:

    ```console
    $ brig project list

    ID           	DESCRIPTION                         	AGE
    first-project	My new Brigade project               	1m
    ```

### Create an Event

With our project defined, we are now ready to manually create an event and watch
Brigade handle it:

```console
$ brig event create --project first-project --follow
```

Below is example output:

```console
Created event "2cb85062-f964-454d-ac5c-526cdbdd2679".

Waiting for event's worker to be RUNNING...
2021-08-10T16:52:01.699Z INFO: brigade-worker version: v2.3.1
2021-08-10T16:52:01.701Z DEBUG: writing default brigade.ts to /var/vcs/.brigade/brigade.ts
2021-08-10T16:52:01.702Z DEBUG: using npm as the package manager
2021-08-10T16:52:01.702Z DEBUG: path /var/vcs/.brigade/node_modules/@brigadecore does not exist; creating it
2021-08-10T16:52:01.702Z DEBUG: polyfilling @brigadecore/brigadier with /var/brigade-worker/brigadier-polyfill
2021-08-10T16:52:01.703Z DEBUG: compiling brigade.ts with flags --target ES6 --module commonjs --esModuleInterop
2021-08-10T16:52:04.210Z DEBUG: running node brigade.js
2021-08-10T16:52:04.360Z [job: hello] INFO: Creating job hello
2021-08-10T16:52:06.921Z [job: hello] DEBUG: Current job phase is SUCCEEDED
```

> ⚠️ By default, Brigade's scheduler scans for new projects every thirty
> seconds. If Brigade is slow to handle your first event, this may be why.

## Cleanup

If you want to keep your Brigade installation, run the following command to
remove the example project created in this QuickStart:

```console
$ brig project delete --id first-project
```

Otherwise, you can remove _all_ resources created in this QuickStart using:

```console
$ helm delete brigade -n brigade
```

## Next Steps

You now know how to install Brigade on a local, development-grade cluster,
define a project, and manually create an event. Continue on to the
[Read Next](/intro/readnext) document where we suggest more advanced topics to
explore.

## Troubleshooting

* [Installation Does Not Complete Successfully](#installation-does-not-complete-successfully)
* [Login command hangs](#login-command-hangs)

### Installation Does Not Complete Successfully

A common cause for failed Brigade deployments is low disk space on the cluster
node. In a local, development-grade cluster on macOS or Windows, this could be
because insufficient disk space is allocated to Docker Desktop, or the space
allocated is nearly full. If this is the case, it should be evident by examining
logs from Brigade's MongoDB or ActiveMQ Artemis pods. If the logs include
messages such as "No space left on device" or "Disk Full!", then you need to
free up disk space and retry the installation. Running `docker system prune` is
one way to recover disk space.

After you have freed up disk space, remove the bad installation, and then retry
using the following commands:

```console
$ helm uninstall brigade -n brigade
$ helm install brigade \
    oci://ghcr.io/brigadecore/brigade \
    --version v2.3.1 \
    --namespace brigade \
    --wait \
    --timeout 300s
```

### Login Command Hangs

If the `brig login` command hangs, check that you included the `--insecure` (or
`-k`) flag. This flag is required because the default configuration utilized by
this QuickStart makes use of a self-signed certificate.
