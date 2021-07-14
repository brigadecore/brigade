---
title: Brigade Overview
description: High-level view of Brigade.
section: intro
aliases: /intro/
weight: 21
---

Brigade is an event-driven scripting tool that runs on Kubernetes. Here's what that means:

- Brigade is for running scriptable automated tasks in the cloud.
- Brigade does not require you to manage host servers.
- Brigade is particularly well suited for CI and CD workloads such as:
  - Automated testing
  - GitHub hook integration
  - Building artifacts and releases
  - Managing deployments
- Brigade ships with its own API server
  - Interactions occur through Brigade directly, via the `brig` CLI or an SDK
  - Operators managing a Brigade installation can still drop down to the K8s API to investigate Brigade resources.
  - The Brigade server is deployed and managed with Helm, thus hosting can occur on any stock Kubernetes system, from Azure to Minikube.
- Brigade handles authn/authz concerns
  - Third-party authentication options include GitHub's OAuth2 identity provider and OIDC-compliant solutions like Azure Active Directory and Google Identity Platform.
  - Administrators of the Brigade system have full control over authorization, including Service Account, Role and User management.
- Brigade uses familiar technologies to make handling events easy:
  - Project developers write scripts in basic JavaScript or TypeScript (no need to learn Node.js)
  - Utilize OCI/Docker images to handle event workloads
  - Project configuration is written in YAML and easily persisted via your preferred VCS

Brigade is designed for teams. It is not a SaaS, nor does it require a legion of domain experts to configure and run it. Instead, it should be easy to install, configure, and maintain.
