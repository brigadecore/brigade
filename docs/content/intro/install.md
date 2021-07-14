---
title: Installing Brigade
description: 'Quick install guide for Brigade'
section: intro
aliases:
  - /install.md
  - /intro/install.md
  - /topics/intro/install.md
---

Brigade is composed of a server, and a command-line tool, brig.
Before you can install Brigade, ensure that you have the [prerequisites](#prerequisites) installed.

* [Install the Brigade CLI](#install-the-brigade-cli)
* [Install the Brigade Server](#install-the-brigade-server)
* [Install Brigade Gateways](#install-brigade-gateways)

## Prerequisites

* A [Kubernetes cluster].  
  Your cluster should be accessible to the source of your event triggers.
  For example, if you want to trigger events from GitHub, the cluster should have a public ip address, and a domain name that resolves to the cluster.
  If you are using Brigade in a local development environment, the [QuickStart] demonstrates how to access a local KinD or Minikube cluster.
* [Helm] CLI v3+ installed.
* [kubectl] CLI installed.
* Free disk space on the cluster nodes.  
  The installation requires sufficient free disk space and will fail if a cluster node disk is nearly full.

[Kubernetes cluster]: https://kubernetes.io/docs/setup/
[Helm]: https://helm.sh/docs/intro/install/
[kubectl]: https://kubernetes.io/docs/tasks/tools/#kubectl

## Install the Brigade CLI

Install the Brigade CLI, brig, by copying the appropriate binary from our releases page into a directory on your machine that is included in your PATH environment variable.

**linux**
```bash
curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.1/brig-linux-amd64
chmod +x /usr/local/bin/brig
```

**macos**
```bash
curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.1/brig-darwin-amd64
chmod +x /usr/local/bin/brig
```

**windows**
```powershell
mkdir -force $env:USERPROFILE\bin
(New-Object Net.WebClient).DownloadFile("https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.1/brig-windows-amd64.exe", "$ENV:USERPROFILE\bin\brig.exe")
$env:PATH+=";$env:USERPROFILE\bin"
```

The script above downloads brig.exe and adds it to your PATH for the current session.
Add the following line to your [PowerShell Profile](https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/) to make the change permanent.

```powershell
$env:PATH+=";$env:USERPROFILE\bin"
```

## Install the Brigade Server

1. Enable Helm's experimental OCI support by setting the HELM_EXPERIMENTAL_OCI environment variable to 1.

    **posix**
    ```bash
    export HELM_EXPERIMENTAL_OCI=1
    ```

    **powershell**
    ```powershell
    $env:HELM_EXPERIMENTAL_OCI=1
    ```

1. Create a directory to store the Brigade Helm charts.

    **posix**
    ```bash
    mkdir -p ~/charts
    ```

    **powershell**
    ```powershell
    mkdir -force $env:USERPROFILE/charts
    ```

1. Run the following commands to install Brigade and wait for it to finish installing.

    ```
    helm chart pull ghcr.io/brigadecore/brigade:v2.0.0-beta.1
    helm chart export ghcr.io/brigadecore/brigade:v2.0.0-beta.1 -d ~/charts
    helm install brigade2 ~/charts/brigade --namespace brigade2 --create-namespace
    kubectl rollout status deployment brigade2-apiserver -n brigade2 --timeout 5m
    ```
   
    If the deployment fails, proceed to the [troubleshooting](#troubleshooting) section.

Now that you have the Brigade server installed, the next step is to [install a Brigade Gateway](#install-brigade-gateways).

### Notes for Azure Kubernetes Services (AKS)

Brigade is well-tested on [Azure Kubernetes Service](https://docs.microsoft.com/en-us/azure/aks/). We recommend using at least Kubernetes 1.6.

- It is recommended to use a Service with type LoadBalancer on AKS, which will generate an Azure load balancer for you.
- For caching and storage, we recommend creating an Azure Storage instance and creating a Persistent Volume and Storage Class that use the `AzureFile` driver.
  For an example, see the `Azure File Setup` section in the [storage document](../../topics/storage/#azure-file-setup).
- You can use Azure Container Registry for private images, provided that you add the ACR instance to the same Resource Group that AKS belongs to.
- ACR's webhooks can be used to trigger events, as they follow the DockerHub webhook format.
- When configuring webhooks, it is recommended that you map a domain via Azure's DNS service, or another DNS service, to your Load Balancer IP.
  GitHub and other webhook services seem to work better with DNS names than with IP addresses.

[QuickStart]: /intro/quickstart/

## Install Brigade Gateways

A Brigade [Gateway] generates events in Brigade in response to external triggers, for example a git push to repository.
Brigade does not install gateways by default, so to complete your Brigade installation, install one or more gateways:

* [Github Gateway](/topics/github/)
* [Container Registry Gateway](/topics/dockerhub/)
* [Generic Gateway](/topics/genericgateway/)

[Gateway]: ../topics/gateways.md

<!--
TODO: Use this in the gateway quickstart
### Brigade Github App Gateway with External IP

Let's take the example of enabling the [GitHub App Gateway](../topics/github.md).

By default, the Brigade Github App gateway chart defines the associated service type as `ClusterIP`, which is only accessible within the Kubernetes cluster.   If we wish to set up the gateway with an externally-visible IP of type `LoadBalancer`, we would upgrade our `brigade-server` release like so:

```
$ helm upgrade brigade-server brigade/brigade --set brigade-github-app.enabled=true --set brigade-github-app.service.type=LoadBalancer
```

We'd then locate the external IP as follows:

```console
$ kubectl get svc brigade-server-brigade-github-app
NAME                                TYPE           CLUSTER-IP    EXTERNAL-IP    PORT(S)          AGE
brigade-server-brigade-github-app   LoadBalancer   10.0.110.59   135.15.52.20   80:30758/TCP     45d
```

(Note that `brigade-server-brigade-github-app` is just the name of the Helm release (`brigade-server`) with `-brigade-github-app` appended)

The `EXTERNAL-IP` field is the IP address that external services, such as GitHub in this example, will use to trigger actions.

There will be more configuration needed for the `brigade-github-app` sub-chart for GitHub events to reach a Brigade project.
See more at [GitHub App Gateway](../topics/github.md).

Note that this is just one way of configuring Brigade to receive inbound connections. Brigade itself does not care how traffic is routed to it. Those with operational knowledge of Kubernetes may wish to use another method of ingress routing.  See the [Ingress](../topics/ingress.md) doc for more information.
-->

## Troubleshooting

### Brigade installation does not finish successfully

A common cause for failed Brigade deployments is either low disk space on the cluster node, or the amount of disk space allocated to Docker Desktop is nearly full.

Check if that is the problem by looking at the logs for Brigade's mongodb and artemis pods.
If the logs include "No space left on device" or "Disk Full!", then you need to free up disk space and retry the installation.
Running `docker system prune` is one way to recover disk space for Docker.

```console
$ kubectl logs brigade2-mongodb-0
...
mkdir: cannot create directory '/bitnami/mongodb/data': No space left on device
```

```console
$ kubectl logs brigade2-artemis-0
...
2021-05-26 17:20:17,865 WARN  [org.apache.activemq.artemis.core.server] AMQ222212: Disk Full! Blocking message production on address 'healthz'. Clients will report blocked.
```

After you have freed up disk space, remove the bad installation, and then retry the installation using the following commands:

```
helm uninstall brigade2 -n brigade2
helm install brigade2 ~/charts/brigade --namespace brigade2
```
