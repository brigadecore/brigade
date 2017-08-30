# Acid at a glance

Acid is a serverless continuous integration (CI) and continuous delivery platform that is
Kubernetes-native. Here's what that means:

- Acid is for running scriptable automated tasks in the cloud.
- Acid does not require you to manage host servers.
- Acid is particularly well suited for CI and CD workloads such as:
  - Automated testing
  - Github hook integration
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

Acid is designed for teams. It is not a SaaS, nor does it require a legion of domain experts to configure and run it. Instead, it should be easy to install, configure, and maintain.
