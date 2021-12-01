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

# The Brigade scripting API

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
is located [here][brigadier] and the npm package page [here][brigadier npm].

To import the library for use in a script, add the following to the top:

```javascript
const brigadier = require('@brigadecore/brigadier')
```

It is considered idiomatic to destructure the library on import:

```javascript
const { events, Job, Group } = require('@brigadecore/brigadier')
```

## Local development

The brigadier library is actually split into two implementations:

* [brigadier] contains nearly no-op implementations of the library's public
  interfaces
* [brigadier-polyfill] contains the logic to actually communicate with a
  Brigade API server; this is the version subsituted by the worker at runtime

By employing this strategy, developers are offered the possibility of running a
Brigade script locally, without consequence, to support
development/troubleshooting efforts.

Let's look at an example.

First, create a new project directory and place the following script contents
into a file named `brigade.js`:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

events.on("brigade.sh/cli", "exec", async event => {
  let job = new Job("my-first-job", "debian:latest", event);
  job.primaryContainer.command = ["echo"];
  job.primaryContainer.arguments = ["My first job!"];
  await job.run();
});

events.process();
```

Then, create a `package.json` file with our brigadier dependency added:

```json
{
  "dependencies": {
    "@brigadecore/brigadier": "^2.0.0-rc.1"
  }
}
```

Next, fetch the brigadier dependency (and in turn, its dependencies):

```console
$ npm install

added 3 packages, and audited 4 packages in 1s

found 0 vulnerabilities
```

Now we're ready to run our Brigade script in a development capacity, using only
the core `brigadier` library: 

```console
$ node brigade.js
No dummy event file provided
Generating a dummy event
No dummy event type provided
Using default dummy event with source "brigade.sh/cli" and type "exec"
Processing the following dummy event:
{
  id: '7eafd0d3-39e9-4341-bd33-7a215e481024',
  source: 'brigade.sh/cli',
  type: 'exec',
  project: { id: '82259392-feea-4102-a8a3-080fdd85cfa9', secrets: {} },
  worker: {
    apiAddress: 'https://brigade2.example.com',
    apiToken: '7000152b-cd0d-483f-b21f-5ef20292e72a',
    configFilesDirectory: '.brigade',
    defaultConfigFiles: {}
  }
}
The Brigade worker would run job my-first-job here.
```

Success!

Say we forgot to add the `events.process()` call at the bottom of our Brigade
script. We'd know immediately when executing the script as there would be no
output at all, signaling that the event handler did not run.

### Optional event config

Developers can optionally provide the following when running their scripts
locally:

  * `BRIGADE_EVENT_FILE` - This is the path to a file containing a JSON
    representation of a dummy event. To see what a valid event looks like from
    Brigadier's perspective, see the dummy event example in the output above
    or refer to [events.ts]
  * `BRIGADE_EVENT` - This is a string of the form `<source>:<type>` to specify
    the event source and type that will be handled by the script. In the
    example above, the dummy event uses `brigade.sh/cli:exec`

For further example usage of brigadier, please review the [Scripting guide]
and/or peruse the [Examples].

[brigadier]: https://github.com/brigadecore/brigade/tree/main/v2/brigadier
[brigadier npm]: https://www.npmjs.com/package/@brigadecore/brigadier
[brigadier-polyfill]: https://github.com/brigadecore/brigade/tree/main/v2/brigadier-polyfill
[Scripting guide]: /topics/scripting/guide
[events.ts]: https://github.com/brigadecore/brigade/tree/main/v2/brigadier/src/events.ts
[Examples]: /topics/examples

## Brigadier API Documentation

Documentation for the Brigadier API is generated from the code directly. It can
be seen in its two forms: generated directly from the TypeScript source code
and generated from the compiled JavaScript.

For the TypeScript documentation, see https://brigadecore.github.io/brigade/ts

For the JavaScript documentation, see https://brigadecore.github.io/brigade/js