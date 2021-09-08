---
title: The Brigade scripting API
description: Describing the public APIs typically used for writing Brigade scripts
section: scripting
weight: 2
aliases:
  - /brigadier.md
  - /topics/brigadier.md
  - /topics/scripting/brigadier.md
---

# The Brigade.js API

This document describes the public APIs typically used for writing Brigade
scripts in either JavaScript or TypeScript. It does not describe internal
libraries, nor does it list non-public methods and properties on these objects.

A Brigade script is executed inside of a [Brigade Worker]. The default worker
contains a Node.js environment for installing [dependencies] and running the
supplied script.

[Brigade Worker]: /topics/scripting/worker
[dependencies]: /topics/scripting/dependencies

## High-level Concepts

An Brigade JS/TS file is always associated with a [project]. A project
defines contextual information, and also dictates the security parameters
under which the script will execute.

A project may associate the script to a git repository. Otherwise, a project's
script must be defined in-line on the project definition.

Brigade files respond to [events]. That is, Brigade scripts are typically
composed of one or more _event handlers_. When the Brigade environment triggers
an event, the associated event handler will be called.

[project]: /topics/project-developers/projects
[events]: /topics/project-developers/events

## The `brigadier` Library

The main library for Brigade is called `brigadier`. The default Brigade Worker
sets up access to this library automatically. The source code for this library
is located [here][brigadier source code] and the npm package page
[here][brigadier npm page].

To import the library for use in a script, add the following to the top:

```javascript
const brigadier = require('@brigadecore/brigadier')
```

It is considered idiomatic to destructure the library on import:

```javascript
const { events, Job, Group } = require('@brigadecore/brigadier')
```

For further example usage of brigadier, please review the [Scripting guide]
and/or peruse the [Examples].

[brigadier source code]: https://github.com/brigadecore/brigade/tree/v2/v2/brigadier
[brigadier npm page]: https://www.npmjs.com/package/@brigadecore/brigadier
[Scripting guide]: /topics/scripting/guide
[Examples]: /topics/examples

## Brigadier API Documentation

Documentation for the Brigadier API is generated from the code directly. It can
be seen in its two forms: generated directly from the TypeScript source code
and generated from the compiled JavaScript.

For the TypeScript documentation, see https://brigadecore.github.io/brigade/ts

For the JavaScript documentation, see https://brigadecore.github.io/brigade/js