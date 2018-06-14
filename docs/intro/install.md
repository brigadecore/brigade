# Quick install guide

_This part is a work-in-progress because Brigade is still developer-oriented_

Brigade is deployed via Helm. Here are the steps:

1. Make sure `helm` is installed, and `helm version` returns the correct server.
2. Add the Brigade repo: `helm repo add brigade https://azure.github.io/brigade`
3. Install Brigade: `helm install brigade/brigade --name brigade-server`

At this point, you have a running Brigade service. You can use `helm get brigade-server` and other Helm tools to examine your running Brigade server.

## Cluster Ingress

By default, Brigade is configured to set up a service as a load balancer for your Brigade build system. To find out your IP address, run:

```console
$ kubectl get svc brigade-server-brigade-github-gw
NAME                               TYPE           CLUSTER-IP    EXTERNAL-IP    PORT(S)          AGE
brigade-server-brigade-github-gw   LoadBalancer   10.0.110.59   135.15.52.20   7744:32394/TCP   45d
```

(Note that `brigade-server-brigade-github-gw` is just the name of the Helm release (`brigade-server`) with `-brigade-github-gw` appended)

The `EXTERNAL-IP` field is the IP address that external services, such as GitHub, will use to trigger actions.

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
- You can use Azure Container Registry for private images, provided that you
  add the ACR instance to the same Resource Group that AKS belongs to.
- ACR's webhooks can be used to trigger events, as they follow the DockerHub
  webhook format.
- When configuring webhooks, it is recommended that you map a domain (via Azure's
  DNS service or another DNS service) to your Load Balancer IP. GitHub and other
  webhook services seem to work better with DNS names than with IP addresses.

---

Prev: [Overview][overview] `|` Next: [Writing your first CI pipeline, Part 1][part1]

[overview]: overview.md
[part1]: tutorial01.md