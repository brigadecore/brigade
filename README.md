# Brigade: Event-based Scripting for Kubernetes

![Build Status](http://badges.technosophos.me/v1/github/build/Azure/brigade/badge.svg?branch=master)

Script simple and complex workflows using JavaScript. Chain together containers,
running them in parallel or serially. Fire scripts based on times, GitHub events,
Docker pushes, or any other trigger. Brigade is the tool for creating pipelines
for Kubernetes.

- JavaScript scripting
- Project-based management
- Configurable event hooks
- Easy construction of pipelines
- Check out the [docs](https://docs.brigade.sh/) to get started.

 <!-- [![asciicast](https://asciinema.org/a/JBsjOpah4nTBvjqDT5dAWvefG.png)](https://asciinema.org/a/JBsjOpah4nTBvjqDT5dAWvefG) -->

## The Brigade Technology Stack

- Brigade :heart: JavaScript: Writing Brigade pipelines is as easy as writing a few lines of JavaScript.
- Brigade :heart: Kubernetes: Brigade is Kubernetes-native. Your builds are translated into
  pods, secrets, and services
- Brigade :heart: Docker: No need for special plugins or elaborate extensions. Brigade uses
  off-the-shelf Docker images to run your jobs. And Brigade also supports DockerHub
  webhooks.
- Brigade :heart: GitHub: Brigade comes with built-in support for GitHub, DockerHub, and
  other popular web services. And it can be easily extended to support your own
  services.

The [design introduction](https://docs.brigade.sh/topics/design/) introduces Brigade concepts and
architecture.

## Quickstart

Check out the quickstart on the docs [here](https://docs.brigade.sh/intro/quickstart/).

## Related Projects

- [Kashti](https://github.com/brigadecore/kashti) - a dashboard for your Brigade pipelines.
- [Brigadeterm](https://github.com/slok/brigadeterm) - a simple terminal ui for brigade pipelining system.
- [Brigade exporter](https://github.com/slok/brigade-exporter) - a [Prometheus](https://prometheus.io) exporter to gather metrics from Brigade.
- Gateways
  - [BitBucket events](https://github.com/lukepatrick/brigade-bitbucket-gateway): Gateway Support for BitBucket repositories
  - [GitLab events](https://github.com/lukepatrick/brigade-gitlab-gateway): Gateway Support for GitLab repositories
  - [Kubernetes events](https://github.com/brigadecore/brigade-k8s-gateway): Gateway that listens to Kubernetes event stream
  - [Event Grid gateway](https://github.com/radu-matei/brigade-eventgrid-gateway): Gateway for Azure Event Grid events
  - [Cron Gateway](https://github.com/technosophos/brigade-cron): Schedule events to run at a particular time
  - [Trello and Generic Webhooks](https://github.com/technosophos/brigade-trello): Experimental gateway for Trello and for generic webhooks
  - [Draft Pack for Building Custom Gateways](https://github.com/technosophos/draft-brigade): Build your own gateway [in 5 minutes](http://technosophos.com/2018/04/23/building-brigade-gateways-the-easy-way.html)
  - [Azure DevOps / VSTS gateway](https://github.com/radu-matei/brigade-vsts-gateway): Gateway for Azure DevOps / VSTS events

## Brigade :heart: Developers

To get started head to the [developer's guide](https://docs.brigade.sh/topics/developers/)

Brigade is well-tested on Minikube and [Azure Kubernetes Service](https://docs.microsoft.com/en-us/azure/aks/).

# Contributing

The Brigade project accepts contributions via GitHub pull requests. The [Contributing](CONTRIBUTING.md) document outlines the process to help get your contribution accepted.


# Support & Feedback

We have a slack channel! [Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free to join for any support questions or feedback, we are happy to help. To report an issue or to request a feature open an issue [here](https://github.com/brigadecore/brigade/issues)

[brigade-project-chart]: https://github.com/brigadecore/charts/tree/master/charts/brigade-project
