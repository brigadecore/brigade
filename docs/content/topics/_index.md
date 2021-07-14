---
title: "Topic Guides"
description: "Deep dives into individual parts of Brigade"
aliases:
  - /topics.md
  - /topics/index.md
  - /topics/index/
---

# Topic Guides

This section of the documentation dives deep into individual parts of Brigade. We'll look at Brigade's architecture,
explore the three main roles of interaction with Brigade, go over Brigade's scripting API and much more.
This is probably where you’ll want to spend most of your time; if you work your way through these guides you
should come out knowing pretty much everything there is to know about Brigade.

If you don't see a topic guide here and have a reasonable level of knowledge on the subject, why not [write one up][write]?

## Table of Contents

- Architecture
  - [Brigade Design](design): A high-level explanation of how Brigade is designed.
- [Operators](operators/_index): An overview of the Operator role in Brigade
  - [Deployment](operators/deploy): How to deploy and manage Brigade
  - [Securing Brigade](operators/security): Securing Brigade via TLS, Ingress and Third-Party Auth
  - [Gateways](operators/gateways): Using and developing Brigade gateways
- [Administrators](administrators/_index): An overview of the Administrator role in Brigade
  - [Authentication](administrators/authentication): Authentication strategies in Brigade
  - [Authorization](administrators/authorization): Authorization setup in Brigade
- [Project Developers](project-developers/_index): An overview of the Project Developer role in Brigade
  - [Projects](project-developers/projects): Creating and managing Brigade Projects
  - [Events](project-developers/events): Understanding and handling Brigade Events
  - [Using Secrets](project-developers/secrets): Using secrets in Brigade Projects
  - [Using the Brigade CLI](project-developers/brig): Using the brig CLI to interact with Brigade
  - [Brigterm](project-developers/brigterm): Using the Brigterm visualization utility
- [Scripting](scripting/_index): Scripting in Brigade
  - [Scripting Guide](scripting/guide): A guide to writing `brigade.js`/`brigade.ts` scripts in Brigade
  - [Brigadier](scripting/brigadier): Brigadier: the JS/TS library for Brigade scripts
  - [Scripting Guide - Advanced](scripting/advanced): Advanced examples for Brigade scripts.
  - [Adding dependencies](scripting/dependencies): How to add local dependencies and NPM packages to your Brigade scripts.
  - [Workers](scripting/workers): More information regarding the Brigade Worker.
- Contributing to Brigade Development
  - [Brigade Developers Guide](developers): A guide for people who want to modify Brigade's code.
- Examples
  - [Example Projects](examples): Example Brigade projects

[write]: https://github.com/brigadecore/brigade/new/master/content/docs/topics
