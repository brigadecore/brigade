# Topic Guides

This section of the documentation dives deep into individual parts of Brigade. This is where complete guides to Brigade's controller, sandbox engine, and much more live. This is probably where you’ll want to spend most of your time; if you work your way through these guides you should come out knowing pretty much everything there is to know about Brigade.

If you don't see a topic guide here and have a reasonable level of knowledge on the subject, why not [write one up][write]?

## Table of Contents

- Architecture
  - [Brigade Design](design.md): A high-level explanation of how Brigade is designed.
- Using Brigade (brigade.js, webhooks)
  - [Scripting Guide](scripting.md): How to write JavaScript for `brigade.js` files.
  - [Brigade.js Reference](javascript.md): The API for brigade.js files.
  - [Scripting Guide - Advanced](scripting_advanced.md): Advanced examples for `brigade.js` files.
  - [GitHub Integration](github.md): A guide for configuring GitHub integration.
  - [Container Registry Integration](dockerhub.md): A guide for configuring integration with DockerHub or Azure Container Registry.
  - [Generic Webhook](genericwebhook.md): How to use Brigade's Generic Webhook functionality.
  - [Using Secrets](secrets.md): How to pass sensitive data into builds.
  - [Brigade Gateways](gateways.md): Learn how to write your own Brigade gateway.
- Configuring and Running Brigade
  - [Projects](projects.md): Install, upgrade, and use Brigade Projects.
  - [Securing Brigade](security.md): Things to consider when configuring Brigade.
  - [Storage](storage.md): How Brigade uses Kubernetes Persistent Storage
  - [Workers](workers.md): More information regarding Brigade Worker
- Contributing to Brigade Development
  - [Brigade Developers Guide](developers.md): A guide for people who want to modify Brigade's
    code.
- Examples
  - [Example Projects](../index.md#technical): Brigade-related projects



[write]: https://github.com/Azure/brigade/new/master/docs/topics
