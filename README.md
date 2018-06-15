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
- Check out the [docs](https://azure.github.io/brigade/) to get started.

[![asciicast](https://asciinema.org/a/JBsjOpah4nTBvjqDT5dAWvefG.png)](https://asciinema.org/a/JBsjOpah4nTBvjqDT5dAWvefG)

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

The [design introduction](https://azure.github.io/brigade/topics/design.html) introduces Brigade concepts and
architecture.

## Quickstart

1. Install Brigade
2. Create a Brigade project
3. Write a Brigade script
4. Execute the script

The easiest way to install Brigade into your Kubernetes cluster is to install it using Helm.

```console
$ helm repo add brigade https://azure.github.io/brigade
$ helm install -n brigade brigade/brigade
```

You will now have Brigade installed.

To create new projects, use the `brigade-project` Helm chart. While inside the Git
repository cloned above, run these commands:

```console
$ helm inspect values brigade/brigade-project > myvalues.yaml
$ # edit myvalues.yaml
```

When editing `myvalues.yaml`, follow the instructions in that file for configuring
your new project. Once you have customized that file, you can install the project
based on your new configuration by passing it with `-f myvalues.yaml`.

```console
$ helm install --name my-project brigade/brigade-project -f myvalues.yaml
```

Now creating your first `brigade.js` is as easy as this:

```javascript
const { events } = require('brigadier')

events.on("exec", (brigadeEvent, project) => {
  console.log("Hello world!")
})
```

Check out [the tutorial](https://azure.github.io/brigade/intro/) for more on creating scripts.

> You can download the latest version of the Brig client from [the releases page](https://github.com/Azure/brigade/releases)

Assuming you named your project `deis/empty-testbed`, you can run a `brigade.js`
file like this:

```console
$ brig run -f brigade.js deis/empty-testbed
```

This will show you the detailed output of running your `brigade.js` script's
`exec` hook.

(To see the names of your projects, run `brig project list`.)

## Related Projects

* [Kashti](https://github.com/Azure/kashti) - a dashboard for your Brigade pipelines.
* [Brigadeterm](https://github.com/slok/brigadeterm) - a simple terminal ui for brigade pipelining system.
* Gateways
  - [BitBucket events](https://github.com/lukepatrick/brigade-bitbucket-gateway): Gateway Support for BitBucket repositories
  - [GitLab events](https://github.com/lukepatrick/brigade-gitlab-gateway): Gateway Support for GitLab repositories
  - [Kubernetes events](https://github.com/azure/brigade-k8s-gateway): Gateway that listens to Kubernetes event stream
  - [Event Grid gateway](https://github.com/radu-matei/brigade-eventgrid-gateway): Gateway for Azure Event Grid events
  - [Cron Gateway](https://github.com/technosophos/brigade-cron): Schedule events to run at a particular time
  - [Trello and Generic Webhooks](https://github.com/technosophos/brigade-trello): Experimental gateway for Trello and for generic webhooks
  - [Draft Pack for Building Custom Gateways](https://github.com/technosophos/draft-brigade): Build your own gateway [in 5 minutes](http://technosophos.com/2018/04/23/building-brigade-gateways-the-easy-way.html)

## Brigade :heart: Developers

To get started head to the [developer's guide](https://azure.github.io/brigade/topics/developers.html)

Brigade is well-tested on Minikube and Azure Container Services.

# Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.
