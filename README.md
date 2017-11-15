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

The [design introduction](docs/topics/design.md) introduces Brigade concepts and
architecture.

## Quickstart

1. Install Brigade
2. Create a Brigade project
3. Write a Brigade script
4. Execute the script

The easiest way to install Brigade into your Kubernetes cluster is to install it using Helm.

```console
$ git clone https://github.com/Azure/brigade.git
$ cd brigade
$ helm install --name brigade ./charts/brigade
```

You will now have Brigade installed.

To create new projects, use the `brigade-project` Helm chart. While inside the Git
repository cloned above, run these commands:

```console
$ helm inspect values ./charts/brigade-project > myvalues.yaml
$ # edit myvalues.yaml
$ helm install --name my-project ./charts/brigade-project -f myvalues.yaml
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

Check out [the tutorial](/docs/intro/) for more on creating scripts.

> In the future, Brigade will provide prebuilt `brig` binaries. But currently you
need to build your own. Take a look at the [Developer's Guide](/docs/topics/developers.md)
to learn more.

Assuming you named your project `deis/empty-testbed`, you can run a `brigade.js`
file like this:

```console
$ brig run -f brigade.js deis/empty-testbed
```

This will show you the detailed output of running your `brigade.js` script's
`exec` hook.

(To see the names of your projects, run `brig project list`.)

## Brigade :heart: Developers

To get started head to the [developer's guide](docs/topics/developers.md)

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
