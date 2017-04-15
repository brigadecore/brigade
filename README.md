# Acid: Acme Continuous Integration and Deployment

Acid is a native Kubernetes CI/CD system. Here's how it works:

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
run({
  name: "run-unit-tests",
  image: "acid-ubuntu:latest",
  tasks:[
    "echo 'running tests...'",
    "make test"
  ]
}, pushRecord);
```

The above creates a new job named `run-unit-tests`. It starts with the AcidIC
image `acid-ubuntu`. And then it runs two tasks:

- `echo 'running tests'` to print a log message
- `make test` to run the project's `Makefile test` target.

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

## Acid :heart: Kubernetes

Acid is Kubernetes-native. Your Acid jobs are translated into one or more Kubernetes
pods, configmaps, and secrets. Acid launches these resources into Kubernetes and
then monitors them for status changes.

The Acid server can run on or off cluster. Your choice.

## Development

To get started:

- Clone this repo
- Run `glide install` to prepare the environment
- Run `make build` to build the source
- Run `bin/acid` to start the server

To build the Docker images, use `make docker-build`.

