---
title: Operators
description: Brigade system operators
section: topics
weight: 3
aliases:
  - /index.md
  - /topics/operators
  - /topics/operators/index.md
---

# Brigade Operators

Users who are in charge of the operational concerns around deploying and
managing the Brigade server and any auxiliary systems are known as
Operators.

In constrast to the other roles of [Administrator] and [Project Developer],
which primarily deal exclusively with Brigade itself, this role requires
knowledge of systems required by Brigade in order to run and function properly,
including:

## Must-knows
Knowledge of these foundational technologies are required to deploy and manage
Brigade.

  * [Kubernetes]: The substrate that hosts Brigade and serves as its runtime environment
  * [Helm]: The tool with which Brigade is packaged, configured and deployed

## Nice-to-haves
These are technologies that Brigade uses under the hood.  While detailed
knowledge shouldn't be required for day-to-day operations, familiarity with
these systems may prove useful.

  * [MongoDB]: Brigade uses Mongo for its data storage layer
  * [ActiveMQ Artemis]: Messaging queue used by Brigade

Next, let's look at how to [deploy Brigade].

[Administrator]: /topics/administrators
[Project Developer]: /topics/project-developers
[Kubernetes]: https://kubernetes.io
[Helm]: https://helm.sh
[MongoDB]: https://www.mongodb.com/
[ActiveMQ Artemis]: https://activemq.apache.org/components/artemis/
[deploy Brigade]: /topics/operators/deploy