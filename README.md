# Brigade: Event-based Scripting for Kubernetes

Script simple and complex workflows using JavaScript. Chain together containers,
running them in parallel or serially. Fire scripts based on times, GitHub events,
Docker pushes, or any other trigger. Brigade is the tool for creating pipelines
for Kubernetes.

- JavaScript scripting
- Project-based management
- Configurable event hooks
- Easy construction of pipelines
- Check out the [docs](/docs/) to get started.

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

## Quickstart

The easiest way to get started with Brigade is to install it using Helm:

```console
$ git clone https://github.com/Azure/brigade.git
$ cd brigade
$ helm install --name brigade ./chart/brigade
```

You will now have Brigade installed.

To create new projects, use the `brigade-project` Helm chart. While inside the Git
repository cloned above, run these commands:

```console
$ helm inspect values ./brigade-project > myvalues.yaml
$ # edit myvalues.yaml
$ helm install --name my-project ./brigade-project -f myvalues.yaml
```

When editing `myvalues.yaml`, follow the instructions in that file for configuring
your new project. Once you have customized that file, you can install the project
based on your new configuration by passing it with `-f myvalues.yaml`.

Now creating your first `brigade.js` is as easy as this:

```javascript
const { events } = require('brigadier')

events.on("exec", (brigadeEvent, project) => {
  console.log("Hello world!")
})
```

But don't be fooled by its simplicty. Brigade can be used to create complex distributed
pipelines. Check out [the tutorial](/docs/intro/) for more.

## `brig`: The Brigade client

Brigade is an event-driven system. Brigade projects live inside of your cluster.
But it's easy to load and run brigade scripts with the `brig` client.

### Building Brig

```
$ make bootstrap build-client
$ bin/brig --help
```

### Running a simple Brig script

Assuming you have a project named `my/project`, you can run a `brigade.js` file like this:

```console
$ brig run -f brigade.js my/project
```

This will show you the detailed output of running your project.

> We suggest starting with the simple `brigade.js` script above, then heading over
to the [docs](/docs/) to learn more.

To see the names of your projects, run `brig project list`.

## Brigade :heart: Developers

These directions assume you are using `minikube` for development. For other environments,
you must make sure you push the Docker images to the right registry or cluster
Docker daemon.

To get started:

- Clone this repo and change directories into it
- Point to MiniKube's Docker environment with `eval $(minikube docker-env)`
- Run `make bootstrap build docker-build` to build the source
- Install the Helm chart: `helm install -n brigade chart/brigade`

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
