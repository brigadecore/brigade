# Acid: Acme Continuous Integration and Deployment

[![Build Status](http://acid.technosophos.me:7744/log/deis/acid/status.svg)](http://acid.technosophos.me:7744/log/deis/acid/id/master)

Acid is a tool for running scriptable automated tasks in the cloud. It is ideally
suited for CI/CD workloads. Acid runs as part of a Kubernetes cluster.

For example, Acid can be used to run CI tasks on a GitHub repository:

- Install Acid into your Kubernetes cluster (if you haven't already)
- Define an `acid.js` file in the root of your GitHub repository.
- Add a GitHub hook pointing to your Acid server
- On each push event (including tagging), Acid runs your `acid.js` file.

## The Acid Technology Stack

- Acid :heart: JavaScript: Writing Acid pipelines is as easy as writing a few lines of JavaScript.
- Acid :heart: Kubernetes: Acid is Kubernetes-native. Your builds are translated into
  pods, secrets, and services
- Acid :heart: Docker: No need for special plugins or elaborate extensions. Acid uses
  off-the-shelf Docker images to run your jobs. And Acid also supports DockerHub
  webhooks.
- Acid :heart: GitHub: Acid comes with built-in support for GitHub, DockerHub, and
  other popular web services. And it can be easily extended to support your own
  services.

## Quickstart

The easiest way to get started with Acid is to install it using Helm:

```console
$ git clone https://github.com/deis/acid.git
$ cd acid
$ helm install --name acid ./chart/acid
```

You will now have Acid installed.

To create new projects, use the `acid-project` Helm chart. While inside the Git
repository cloned above, run these commands:

```console
$ helm inspect values ./acid-project > myvalues.yaml
$ # edit myvalues.yaml
$ helm install --name my-project ./acid-project -f myvalues.yaml
```

When editing `myvalues.yaml`, follow the instructions in that file for configuring
your new project. Once you have customized that file, you can install the project
based on your new configuration by passing it with `-f myvalues.yaml`.

Now creating your first `acid.js` is as easy as this:

```javascript
const { events } = require('libacid')

events.on("push", (acidEvent, project) => {
  console.log("Hello world!")
})
```

But don't be fooled by its simplicty. Acid can be used to create complex distributed
pipelines. Check out [the tutorial](/docs/intro/) for more.

## Acid :heart: Developers

To get started:

- Clone this repo
- Run `glide install` to prepare the environment
- Run `make build` to build the source
- Run `bin/acid` to start the server

To build the Docker images, use `make docker-build`.

To deploy via [Draft](https://github.com/Azure/draft), use `make build-docker-bin && draft up`.
