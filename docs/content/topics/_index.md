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
  - [Brigade Design](design): A high-level explanation of how Brigade is designed. (TODO: needs updating)
- Operators
  - Overview/Definition of this role (TODO)
  - Install/Manage system (TODO)
  - Securing Brigade (TLS, ingress, domain, Third-Party Auth) (TODO; see `security.md`)
  - Gateways (TODO; see `gateways.md`)
- Administrators
  - Overview/Definition of this role (TODO)
  - SA/User/Role Management (TODO)
- Project Developers
  - Overview/Definition of this role (TODO)
  - Projects (brig init!) (TODO; see `projects.md`)
  - Events (TODO)
  - Brigterm (when available) (TODO)
- Scripting
  - TODO; here are the links carried over from v1, all of which need updating/editing:
  - [Scripting Guide](scripting): How to write JavaScript for `brigade.js` files.
  - [Using the Brigade CLI](brig)
  - [Brigade.js Reference](javascript): The API for brigade.js files.
    - Note(vadice): I wonder if the inline code documentation is sufficient to point to?
  - [Scripting Guide - Advanced](scripting_advanced): Advanced examples for `brigade.js` files.
  - [Adding dependencies to `brigade.js`](dependencies): How to add local dependencies and NPM packages to your `brigade.js` files.
- Contributing to Brigade Development
  - [Brigade Developers Guide](developers): A guide for people who want to modify Brigade's
    code. (TODO: needs updating)
- Examples
  - [Example Projects](../../../examples): Example Brigade projects


TODO: Misc leftover from v1 to consider keeping/updating:

  - [Workers](workers): More information regarding Brigade Worker.
  - [Using Secrets](secrets): How to pass sensitive data into builds.

[write]: https://github.com/brigadecore/brigade/new/master/content/docs/topics
