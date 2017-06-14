# Acid: Acme Continuous Integration and Deployment

[![Build Status](http://localhost:7744/log/deis/acid/status.svg)](http://localhost:7744/log/deis/acid)

Acid is designed to give you the ease of use of a hosted CI/CD solution, but
running on your own Kubernetes cluster. And it's flexibile enough to accomodate
sophisticated multi-step builds.

Here's how it works:

- Install Acid into your Kubernetes cluster (if you haven't already)
- You define an `acid.js` file in the root of your GitHub repository.
- Add a GitHub hook pointing to your Acid server
- On each push event (including tagging), Acid runs your `acid.js` file.

## Acid :heart: JavaScript

> No More YAML!
> No More JSON!
> Give us a real scripting language!

Acid runs each build in an isolated JavaScript runtime. With simple primitives like
`jobs` and `tasks`, you can organize your Acid scripts to do what you want. Define
however many stages, jobs, and tasks you want.

A simple `acid.js` file looks like this:

```javascript

// Acid lets you respond to different Github events:
events.github.push = function(e) {
  // Define a build step:
  j = new Job("run-unit-tests");

  // Use a custom image (this is actually the default)
  j.image = "acid-ubuntu:latest";

  // Define a couple of tasks:
  j.tasks = [
    "echo 'running tests'",
    "make test"
  ];

  // Run the build:
  j.run()
}
```

The above creates a new job named `run-unit-tests`. It starts with the AcidIC
image `acid-ubuntu:latest` (the default image). And then it runs two tasks:

- `echo 'running tests'` to print a log message
- `make test` to run the project's `Makefile test` target.

Check your `acid.js` file into the root of your project's repository.

## Acid :heart: Kubernetes

Acid is Kubernetes-native. Your Acid jobs are translated into one or more Kubernetes
pods, configmaps, and secrets. Acid launches these resources into Kubernetes and
then monitors them for status changes.

The easiest way to get started with Acid is to install it using Helm:

```console
$ helm repo add acid https://deis.github.io/acid
$ helm install acid/acid
```

To create new products, use the `acid-project` Helm chart:

```console
$ helm fetch acid/acid-project
$ helm inspect values acid-project-*.tgz > myvalues.yaml
$ # edit myvalues.yaml
$ helm install acid-project-*.tgz -f myvalues.yaml
```

_Make sure you change the `secret`_. You will use that secret when setting up GitHub
hooks.

## Acid :heart: GitHub

To add Acid support to your GitHub project, set up an acid server, and then in
your GitHub project:

- Go to "Settings"
- Click "Webhooks"
- Click the "Add webhook" button
- For "Payload URL", add the URL: "http://YOUR_HOSTNAME:7744/events/github"
- For "Content type", choose "application/json"
- For "Secret", use the secret you configured in your Helm config.
- Choose "Just the push event"

## Acid :heart: Docker

Acid runs your builds inside of Docker containers. We created some basic IC
(integration container) images for you. But it's really easy to create your own.
A basic Ruby AcidID `Dockerfile` might look like this:

```Dockerfile
FROM ruby:2.1

COPY rootfs/hook.sh /hook.sh
CMD /hook.sh
```

(We provide a nice little `hook.sh` script to bootstrap your environment, but you
can definitely create your own).

## Acid :heart: Developers

To get started:

- Clone this repo
- Run `glide install` to prepare the environment
- Run `make build` to build the source
- Run `bin/acid` to start the server

To build the Docker images, use `make docker-build`.

