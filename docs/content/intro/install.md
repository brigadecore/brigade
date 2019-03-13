---
title: Installing Brigade
description: 'Quick install guide for Brigade'
section: intro
---

_This part is a work-in-progress because Brigade is still developer-oriented_

The Brigade server is deployed via its [Helm](https://github.com/helm/helm) chart and
Brigade projects are managed via [brig](#brig). Here are the steps:

1. Make sure `helm` is installed, and `helm version` returns the correct server.
2. Add the Brigade repo: `helm repo add brigade https://azure.github.io/brigade-charts`
3. Install Brigade: `helm install brigade/brigade --name brigade-server`
4. Create a Brigade project: `brig project create`

At this point, you have a running Brigade service. You can use `helm get brigade-server` and other Helm tools to examine your running Brigade server.

## Cluster Ingress

By default, Brigade is not configured with a load balancer service for incoming requests.  Rather, cluster ingress
comes in the form of one or more [Gateways](../topics/gateways.md) that provide configurable services, usually in tandem
with ingress resources.

Let's take the example of enabling the [GitHub App Gateway](../topics/github.md).

We would upgrade our `brigade-server` release like so:

```
$ helm upgrade -n brigade-server brigade/brigade --set brigade-github-app.enabled=true
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

Note that this is just one way of configuring Brigade to receive inbound connections. Brigade itself does not care how traffic is routed to it. Those with operational knowledge of Kubernetes may wish to use another method of ingress routing.

## Brig

We recommend using [Brig](https://github.com/Azure/brigade/tree/master/brig), a command line tool for interacting with Brigade. Read the [Brig guide](https://github.com/Azure/brigade/tree/master/brig) for installation and usage docs.

## Notes for Minikube

You can run Brigade on [Minikube](https://github.com/kubernetes/minikube) for easy testing
and development. Minikube provides built-in support for caching and sharing files during
builds. However, there are a few things that are much harder to do when running locally:

- Listening for GitHub webhooks requires you to route inbound traffic from the internet
  to your Minikube cluster. We do not recommend doing this unless you really understand
  what you are doing.
- Other inbound services may also be limited by the same restriction.

## Notes for Azure Container Services (AKS)

Brigade is well-tested on [AKS Kubernetes](https://docs.microsoft.com/en-us/azure/aks/). We recommend using at least Kubernetes 1.6.

- It is recommended to use a Service with type LoadBalancer on AKS, which will generate
  an Azure load balancer for you.
- For caching and storage, we recommend creating an Azure Storage instance and
  creating a Persistent Volume and Storage Class that use the `AzureFile` driver.
  (For an example, see the `Azure File Setup` section in the [storage document](../topics/storage.md#azure-file-setup).)
- You can use Azure Container Registry for private images, provided that you
  add the ACR instance to the same Resource Group that AKS belongs to.
- ACR's webhooks can be used to trigger events, as they follow the DockerHub
  webhook format.
- When configuring webhooks, it is recommended that you map a domain (via Azure's
  DNS service or another DNS service) to your Load Balancer IP. GitHub and other
  webhook services seem to work better with DNS names than with IP addresses.

[overview]: ../overview
[part1]: ../tutorial01