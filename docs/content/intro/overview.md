---
title: Brigade Overview
description: High level view of the Brigade tool.
section: intro
aliases: 
  - /overview.md
  - /intro/overview.md
---

Brigade is a Kubernetes-native tool for doing event-driven scripting. Here's what that means:

- Brigade is for running scriptable automated tasks in the cloud.
- Brigade does not require you to manage host servers.
- Brigade is particularly well suited for CI and CD workloads such as:
  - Automated testing
  - GitHub hook integration
  - Building artifacts and releases
  - Managing deployments
- Brigade is built directly on Kubernetes APIs, which means...
  - You can deploy Brigade onto any stock Kubernetes system, from Azure to Minikube
  - You can monitor Brigade and its jobs using Kubernetes tools (or with Brigade's own tools, of course)
  - Brigade uses Kubernetes resource types
  - Brigade can be deployed and managed with `helm`.
  - Brigade can easily be integrated with `draft`.
- Brigade uses technologies designed to make builds easy:
  - Write configuration in basic JavaScript (no need to learn Node.js)
  - Encapsulate build steps in Docker images. Or better yet, just use off-the-shelf
    Docker images.

Brigade is designed for teams. It is not a SaaS, nor does it require a legion of domain experts to configure and run it. Instead, it should be easy to install, configure, and maintain.
