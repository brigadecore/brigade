# Acid Design

This is a high-level explanation of the design of Acid.

Acid is a Kubernetes-native CI/CD system that ties Git repositories to a build
pipeline.

## Terminology

- **Acid** is the main server. It is designed to run in-cluster as a deployment.
- **AcidIC**: Acid _integration container_ is a container for running a job.
- **Acid.js**: A JavaScript file that contains an Acid configuration. Pronounced
  "Acid Jay Es" or "Acid Jazz".
- **Job**: A build unit, comprised of one or more build steps called "tasks"
- **Webhook**: An incoming request from an external SCM that provides a JSON
  payload indicating that a new contribution has been made. This triggers an
  Acid run.

## Components

Acid is composed of the following pieces:

- The Acid server
- AcidIC images (acid-ubuntu, acid-go,...)
- The Acid.js supporting libraries (runner.js, quokka)
- The Acid Helm chart (installs Acid into Kubernetes)
- The Acid Project Helm Chart (add or manage a project in an Acid server)

## High-Level Flow

When a GitHub project is configured to send webhooks to Acid, it will send one
hook request per `push` event.

![Flow Diagram](sequence.png)

A hook kicks of an Acid build, which in turn will invoke the repository's `acid.js` file.
The build is done inside of Kubernetes, with each `Job` being run as a Kubernetes
pod.

## The Server

The Acid server is a _webhook provider_. It listens for GitHub webhook requests
on port _7744_ (default), and manages builds.

When a Webhook `push` event is triggered, Acid will do the following:

- Load the data provided by the webhook
- Load the project configuration.
  - A configuration is stored in a Kubernetes secret.
  - A configuration has a name, a repository URL, and a GitHub secret (for Auth)
  - If a request comes in for a project that does not match any of the configurations,
    Acid returns an error.
- Perform auth against the original payload
  - AuthN is done using GitHub's hook auth mechanism -- a cryptographic hash with
    an agreed-upon secret salt.
  - If auth fails, Acid returns an error
- Clone the GitHub repo (or update if the repo is cached)
- Find and load the acid.js file
  - acid.js must be at the repository root
  - if no file is found, Acid returns an error
- Prepare the JavaScript runtime (sandboxed; one per request; never re-used)
- Run the acid.js file
  - For each Acid `Job`, create a config map and a pod.
  - The config map stores instructions on what to execute.
  - The pod mounts the config map as a volume
  - Run until the script is complete. Most jobs are blocking, but jobs can be
    run in the background.
  - On error, return an error to GitHub
  - On success, return 200 OK to GitHub

This is the basic operation of the Acid webhook server.

## Acid IC (Integration Containers)

An acid.js file may specify which Docker image to run as part of a build step.
These images should have a specific set of traits that mark them as integration
containers. Primarily, they must execute the instructions passed by the Acid
server.

## JavaScript and acid.js

Acid has a full JavaScript engine inside. This engine provides some supporting
libraries to provide primitives for:

- Creating and managing jobs
- Accessing configuration
- Querying Kubernetes
- Performing basic concurrency tasks

It does _not_ allow loading of external JS via NPM or other JavaScript loaders.

## Kubernetes Objects

Acid defines itself in terms of the following Kubernetes objects:

The Acid server runs (preferably) as a Deployment inside of Kubernetes.

Projects map GitHub projects to Acid build tasks. These are stored inside of
Kubernetes secrets since they contain sensitive information.

When a Job object is created in acid.js, this translates to a pod and a configmap.
The pod is configured to run exactly once. It mounts the configmap as a volume.
Optionally, an acid.js author may expose some of the items in the project's
secret to the pod. These are supplied as environment variables.


