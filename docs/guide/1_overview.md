# Overview of Acid

Acid is a serverless continuous integration (CI) and continuous delivery platform that is
Kubernetes-native. Here's what that means:

- Acid is for running scriptable automated tasks in the cloud.
- Acid does not require you to manager host servers.
- Acid is particularly well suited for CI and CD workloads such as:
  - Automated testing
  - GitHub hook integration
  - Building artifacts and releases
  - Managing deployments
- Acid is built directly on Kubernetes APIs, which means...
  - You can deploy Acid onto any stock Kubernetes system, from Azure to Minikube
  - You can monitor Acid and its jobs using Kubernetes tools (or with Acid's own tools, of course)
  - Acid uses Kubernetes resource types
  - Acid can be deployed and managed with `helm`.
  - Acid can easily be integrated with `draft`.
- Acid uses technologies designed to make builds easy:
  - Write configuration in basic JavaScript (no need to learn Node.js)
  - Encapsulate build steps in Docker images. Or better yet, just use off-the-shelf
    Docker images.

Acid is designed for teams. It is not a SaaS, nor does it require a legion of
domain experts to configure and run it. Instead, it should be easy to install,
configure, and maintain.

## High-Level Flow

This section provides an overview of how Acid processes requests.

At a high level, Acid can handle different sorts of requests. To provide a simple example, here is a GitHub workflow:

![Acid Webhook Flow](../Acid-webhook.png)

GitHub hosts a number of projects. Our team has configured two Acid projects (`github.com/technosophos/example` and `github.com/helm/otherexample`). Likewise, GitHub has been configured to trigger an Acid build, via webhook, whenever a pull request is received.

1. Event: Github sends a webhook to Acid. Acid authenticates the request.
2. Load Config: Acid loads the configuration for the given GitHub repository. This configuration
   may include credentials, special configuration directives, and settings or properties for the
   build.
3. Run: Acid fetches the github repository, reads the `acid.js` file, and then executes it. In the
  typical build scenario, this script will invoke one or more jobs (Kubernetes pods) that will build
  and test the code.
4. Notify: When the build is complete, Acid will notify GitHub over the GitHub status API. It will
  send GitHub the state (success, failure, etc) along with a link where the user can fetch the logs

The workflow above can be expressed as a series of events.
When a GitHub project is configured to send webhooks to Acid, it will send one
hook request per `push` event.
