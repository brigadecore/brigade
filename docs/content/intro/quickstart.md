---
title: A Brigade Quickstart
description: A Brigade Quickstart.
section: intro
weight: 3
aliases:
  - /quickstart.md
  - /intro/quickstart.md
---

In this QuickStart, you will install Brigade, create a project and execute it.

* [Prerequisites](#prerequisites)
* [Install Brigade](#install-brigade)
* [Log in to Brigade](#log-in-to-brigade)
* [Create a Project](#create-a-project)
* [Trigger an Event](#trigger-an-event)

## Prerequisites

* [A development Kubernetes 1.16.0+ cluster](#create-a-cluster).
* [Brigade CLI](#install-the-brigade-cli) installed.
* [Helm] v3.7.0+ installed.
* [kubectl] CLI installed.
* Free disk space. The installation requires sufficient free disk space and will fail if your disk is nearly full.

> Please take note that the default configuration is not secure and is not appropriate for any shared cluster.
> This configuration is appropriate for evaluating Brigade on a local development cluster, and should not be used in production.

### Create a Cluster

If you do not already have a development cluster, we recommend using [KinD].
KinD runs a Kubernetes cluster locally using [Docker].
[Minikube] also works well for local development.

1. Ensure that Docker is running on your machine. For example, `docker ps` should return a listing of running containers.

1. Install [KinD]. See the KinD documentation for full installation instructions, below are instructions for common environments:

    **linux**
    ```bash
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
    chmod +x ./kind
    mv ./kind /usr/local/bin
    ```

    **macos with Homebrew**
    ```bash
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64
    chmod +x ./kind
    mv ./kind /usr/local/bin
    ```

    **windows**
    ```powershell
    mkdir -force $env:USERPROFILE\bin
    (New-Object Net.WebClient).DownloadFile("https://kind.sigs.k8s.io/dl/v0.11.1/kind-windows-amd64", "$ENV:USERPROFILE\bin\kind.exe")
    $env:PATH+=";$env:USERPROFILE\bin"
    ```

    The script above downloads kind.exe and adds it to your PATH for the current session.
    Add the following line to your [PowerShell Profile](https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/) to make the change permanent.

    ```powershell
    $env:PATH+=";$env:USERPROFILE\bin"
    ```

1. Create a Kubernetes cluster by running the following command:
    ```
    kind create cluster
    ```

1. Verify that you can connect to the cluster using kubectl:
    ```
    kubectl cluster-info
    ```

[Helm]: https://helm.sh/docs/intro/install/
[Minikube]: https://minikube.sigs.k8s.io/docs/start/
[KinD]: https://kind.sigs.k8s.io/docs/user/quick-start/
[kubectl]: https://kubernetes.io/docs/tasks/tools/#kubectl
[Docker]: https://docs.docker.com/get-docker/

### Install the Brigade CLI

Install the Brigade CLI, brig, by copying the appropriate binary from our releases page into a directory on your machine that is included in your PATH environment variable.

**linux**
```bash
curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.3/brig-linux-amd64
chmod +x /usr/local/bin/brig
```

**macos**
```bash
curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.3/brig-darwin-amd64
chmod +x /usr/local/bin/brig
```

**windows**
```powershell
mkdir -force $env:USERPROFILE\bin
(New-Object Net.WebClient).DownloadFile("https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.3/brig-windows-amd64.exe", "$ENV:USERPROFILE\bin\brig.exe")
$env:PATH+=";$env:USERPROFILE\bin"
```

The script above downloads brig.exe and adds it to your PATH for the current session.
Add the following line to your [PowerShell Profile](https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/) to make the change permanent.

```powershell
$env:PATH+=";$env:USERPROFILE\bin"
```

## Install Brigade

Install Brigade on your local development cluster. See our [Installation] instructions for full instructions suitable for production clusters.

1. Enable Helm's experimental OCI support by setting the `HELM_EXPERIMENTAL_OCI` environment variable to 1.

    **posix**
    ```bash
    export HELM_EXPERIMENTAL_OCI=1
    ```

    **powershell**
    ```powershell
    $env:HELM_EXPERIMENTAL_OCI=1
    ```

1. Run the following commands to install Brigade.

    ```
    helm install brigade \
      oci://ghcr.io/brigadecore/brigade \
      --version v2.0.0-beta.3 \
      --create-namespace \
      --namespace brigade
    kubectl rollout status deployment brigade-apiserver -n brigade --timeout 5m
    ```
    
    Wait for the Brigade deployment to be ready.
    If the deployment fails, proceed to the [installation troubleshooting](/intro/install/#troubleshooting) section.

Now that Brigade is running, you need to determine the address of the Brigade API so that you can use it later in this QuickStart:

### Port Forward a Local Cluster

If you are running a cluster locally, use port forwarding to make the Brigade API available via localhost:

**posix**

```
kubectl --namespace brigade port-forward service/brigade-apiserver 8443:443 &>/dev/null &
```

**powershell**

```
& kubectl --namespace brigade port-forward service/brigade-apiserver 8443:443 *> $null  
```

### Get External IP of a Remote Cluster

If you are running a cluster remotely, such as on a cloud provider, the Brigade API is available at the External IP of the brigade-apiserver service:

```
kubectl get service --namespace brigade brigade-apiserver -o=jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

[Installation]: /intro/install/

## Log in to Brigade

Authenticate to Brigade as the root user. First, you need to acquire the auto-generated root user password using the following kubectl command:

```
export APISERVER_ROOT_PASSWORD=$(kubectl get secret --namespace brigade brigade-apiserver -o=jsonpath='{.data.root-user-password}' | base64 --decode)
```

**local clusters**

```
brig login --insecure --server https://localhost:8443 --root --password "${APISERVER_ROOT_PASSWORD}"
```

The \--insecure flag instructs Brigade to ignore the self-signed certificate used by our local installation of Brigade.

If the address https://localhost:8443 does not resolve, double-check that the brigade-apiserver service was successfully forwarded from the previous section.

**remote clusters**

Replace `IP_ADDRESS` with the External IP address of your cluster:

```
brig login --insecure --server https://IP_ADDRESS --root --password "${APISERVER_ROOT_PASSWORD}"
```

## Create a Project

A Brigade [project] defines event handlers, such as the definition of a CI pipeline.

1. Initialize a new Brigade project with the `brig init` CLI command.

    ```console
    brig init --id first-project
    ```
    This will create a project definition like the following in `.brigade/project.yaml`, which defines a handler for the "exec" event and echoes "Hello, World!".

    ```yaml
    apiVersion: brigade.sh/v2-beta
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

1. Create the project in Brigade with the following command.

    ```console
    brig project create --file .brigade/project.yaml
    ```

1. List the defined projects with `brig project list` and verify that you see your new project:

    ```console
    brig project list
    ID           	DESCRIPTION                         	AGE
    first-project	My new Brigade project               	49m
    ```

[project]: /topics/projects/#an-introduction-to-projects

## Trigger an Event

With our project defined, you are now ready to trigger an event and watch your handler execute.

```
brig event create --project first-project --follow
```

Below is example output of a successful event handler:
```
Created event "2cb85062-f964-454d-ac5c-526cdbdd2679".

Waiting for event's worker to be RUNNING...
2021-08-10T16:52:01.699Z INFO: brigade-worker version: v2.0.0-beta.3
2021-08-10T16:52:01.701Z DEBUG: writing default brigade.ts to /var/vcs/.brigade/brigade.ts
2021-08-10T16:52:01.702Z DEBUG: using npm as the package manager
2021-08-10T16:52:01.702Z DEBUG: path /var/vcs/.brigade/node_modules/@brigadecore does not exist; creating it
2021-08-10T16:52:01.702Z DEBUG: polyfilling @brigadecore/brigadier with /var/brigade-worker/brigadier-polyfill
2021-08-10T16:52:01.703Z DEBUG: compiling brigade.ts with flags --target ES6 --module commonjs --esModuleInterop
2021-08-10T16:52:04.210Z DEBUG: running node brigade.js
2021-08-10T16:52:04.360Z [job: hello] INFO: Creating job hello
2021-08-10T16:52:06.921Z [job: hello] DEBUG: Current job phase is SUCCEEDED
```

## Cleanup

If you want to keep your Brigade installation, run the following command to remove the example project created in this QuickStart:

```
brig project delete --id first-project
```

Otherwise, you can remove ALL resources created in this QuickStart by either:

* Deleting the KinD cluster that you created at the beginning with `kind delete cluster --name kind-kind` OR
* Preserving the cluster and uninstalling Brigade with `helm delete brigade -n brigade`

## Next Steps

You now know how to install Brigade on a local development cluster, define a project, and trigger an event for the project.
Continue on to the [Read Next] document where we review more advanced topics to delve into.

[Read Next]: /intro/readnext

## Troubleshooting

* [Brigade installation does not finish successfully](/intro/install/#troubleshooting)
* [Login command hangs](#login-command-hangs)

### Login command hangs

If the brig login command hangs, check that you included the -k flag.
This flag is required because our local development installation of Brigade is using a self-signed certificate.
