---
title: Brigade Roles
description: An overview of the roles in Brigade
section: topics
weight: 2
aliases:
  - /roles.md
  - /topics/roles.md
---

In this section, we'll look at the three main roles of concern in a
production-grade Brigade deployment. They are essentially divided up according
to the scope of interaction within Brigade:

  * Management of the deployment of the Brigade server and any auxiliary systems
  * Management of users/accounts within Brigade
  * Management and development of Brigade projects.

Naturally, there may be overlap between these roles, and for development
setups, one or two users might cover all of them.  However, they also serve
nicely as categories of context for documentation, so without further ado,
let's dive in.

  * [Operators] - Users who install and manage the deployment of Brigade and gateways
    - [Deployment](operators/deploy): How to deploy and manage Brigade
    - [Securing Brigade](operators/security): Securing Brigade via TLS, Ingress and Third-Party Auth
    - [Storage](operators/storage): Storage options and configuration for Brigade
    - [Gateways](operators/gateways): Using and developing Brigade gateways
  * [Administrators] - Users who manage authentication and authorization concerns in Brigade
    - [Authentication](administrators/authentication): Authentication strategies in Brigade
    - [Authorization](administrators/authorization): Authorization setup in Brigade
  * [Project Developers] - Users who create projects and write Brigade scripts for handling events
    - [Projects](project-developers/projects): Creating and managing Brigade Projects
    - [Events](project-developers/events): Understanding and handling Brigade Events
    - [Using Secrets](project-developers/secrets): Using secrets in Brigade Projects
    - [Using the Brigade CLI](project-developers/brig): Using the brig CLI to interact with Brigade
    - [Brigterm](project-developers/brigterm): Using the Brigterm visualization utility

[Operators]: /operators
[Administrators]: /administrators
[Project Developers]: /project-developers