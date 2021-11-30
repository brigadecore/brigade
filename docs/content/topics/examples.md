---
title: Examples
description: Brigade examples
section: topics
weight: 8
aliases:
  - /examples.md
  - /topics/examples.md
---

# Brigade Examples

Looking for ready-to-run Brigade project examples to use and/or learn from?
Check out the [Examples] directory at the root of the Brigade repo.

[Examples]: https://github.com/brigadecore/brigade/tree/main/examples
## Projects

Here are some of the project types you'll find inside:

  * [First Job][first-job]: A project with an embedded script featuring an
    event handler that creates and runs a Job
  * [Groups][groups]: A project with an embedded script featuring organization
    of multiple Jobs running in sequence and concurrently
  * [Git][git]: A project with a script located in its associated Git
    repository
  * [Shared Workspace][shared-workspace]: A project with an embedded script
    demonstrating use of a shared workspace between a Worker and its Jobs
  * [First Payload][first-payload]: A project with an embedded script
    demonstrating use of the payload from a handled event

[first-job]: https://github.com/brigadecore/brigade/tree/main/examples/03-first-job
[groups]: https://github.com/brigadecore/brigade/tree/main/examples/05-groups
[git]: https://github.com/brigadecore/brigade/tree/main/examples/06-git
[shared-workspace]: https://github.com/brigadecore/brigade/tree/main/examples/10-shared-workspace
[first-payload]: https://github.com/brigadecore/brigade/tree/main/examples/12-first-payload

## Gateways

The [gateways] directory currently has an [example gateway] written using the
Go SDK.

For more context and a description of this example gateway, see the [Gateways]
doc.

[gateways]: https://github.com/brigadecore/brigade/tree/main/examples/gateways
[example gateway]: https://github.com/brigadecore/brigade/tree/main/examples/gateways/example-gateway
[Gateways]: /topics/operators/gateways

