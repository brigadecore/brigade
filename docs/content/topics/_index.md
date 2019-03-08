---
title: 'Topic Guides'
description: 'Deep dives into individual parts of Brigade'
---

# Topic Guides

This section of the documentation dives deep into individual parts of Brigade. This is where complete guides to Brigade's controller, sandbox engine, and much more live. This is probably where you’ll want to spend most of your time; if you work your way through these guides you should come out knowing pretty much everything there is to know about Brigade.

If you don't see a topic guide here and have a reasonable level of knowledge on the subject, why not [write one up][write]?

## Table of Contents

- Architecture
  - [Brigade Design](design): A high-level explanation of how Brigade is designed.
- Using Brigade (brigade.js, webhooks)
  - [Scripting Guide](scripting): How to write JavaScript for `brigade.js` files.
  - [Brigade.js Reference](javascript): The API for brigade.js files.
  - [Scripting Guide - Advanced](scripting_advanced): Advanced examples for `brigade.js` files.
  - [Adding dependencies to `brigade.js`](dependencies): How to add local dependencies and NPM packages to your `brigade.js` files.
  - [GitHub Integration](github): A guide for configuring GitHub integration.
  - [Container Registry Integration](dockerhub): A guide for configuring integration with DockerHub or Azure Container Registry.
  - [Generic Gateway](genericgateway): How to use Brigade's Generic Gateway functionality.
  - [Using Secrets](secrets): How to pass sensitive data into builds.
  - [Brigade Gateways](gateways): Learn how to write your own Brigade gateway.
- Configuring and Running Brigade
  - [Projects](projects): Install, upgrade, and use Brigade Projects.
  - [Securing Brigade](security): Things to consider when configuring Brigade.
  - [Storage](storage): How Brigade uses Kubernetes Persistent Storage.
  - [Workers](workers): More information regarding Brigade Worker.
  - [Testing Brigade](testing): How to test Brigade.
- Contributing to Brigade Development
  - [Brigade Developers Guide](developers): A guide for people who want to modify Brigade's
    code.
- Examples
  - [Example Projects](../index/#technical): Brigade-related projects.



[write]: https://github.com/Azure/brigade/new/master/content/docs/topics