# Brigadier: The JS/TS library for Brigade

Brigadier is the library for writing [Brigade](https://brigade.sh) scripts.

It provides core Brigade objects such as Events and Jobs that developers
can use to construct expressive event-handling logic for their Brigade
projects.

Brigadier supports writing scripts in either JavaScript or TypeScript.

## Installation

[![NPM](https://nodei.co/npm/@brigadecore/brigadier.png)](https://www.npmjs.com/package/@brigadecore/brigadier)

Normally, the brigadier dependency is declared at the top of a Brigade script.
The library itself is pre-loaded in the Brigade Worker:

```javascript
const { events, Job } = require("@brigadecore/brigadier");
```

To facilitate script development, you may also install brigadier to your
environment with Yarn, NPM, etc.:

```sh
$ yarn add @brigadecore/brigadier
```

## Versioning

The 0.x Brigadier npm releases are compatible with
[Brigade v1.x and under](https://github.com/brigadecore/brigade/tree/master).

The 2.x brigadier npm releases are compatible with
[Brigade v2.x](https://github.com/brigadecore/brigade/tree/v2).

## Usage

> Note: the following examples are using brigadier 2.x, compatible with Brigade v2.

Here is an example `brigade.js` script which declares an event handler for
GitHub push events, running tests for the project it is associated with:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

const localPath = "/workspaces/brigade";

events.on("brigade.sh/github", "push", async event => {
  let test = new Job("test", "golang:1.17", event);
  test.primaryContainer.sourceMountPath = localPath;
  test.primaryContainer.workingDirectory = localPath;
  test.primaryContainer.command = ["make"];
  test.primaryContainer.arguments = ["test"];

  await test.run();
})

events.process();
```

Or, the same can be written in TypeScript (`brigade.ts`):

```typescript
import { events, Event, Job } from "@brigadecore/brigadier"

const localPath = "/workspaces/brigade"

events.on("brigade.sh/github", "push", async event => {
  let test = new Job("test", "golang:1.17", event)
  test.primaryContainer.sourceMountPath = localPath
  test.primaryContainer.workingDirectory = localPath
  test.primaryContainer.command = ["make"]
  test.primaryContainer.arguments = ["test"]

  await test.run()
})

events.process()
```

To learn more, visit the [official scripting guide] or explore the Brigade
[example projects].

[official scripting guide]: https://docs.brigade.sh/topics/scripting
[example projects]: https://github.com/brigadecore/brigade/tree/v2/examples

## Contributing

The Brigade project accepts contributions via GitHub pull requests. The
[Contributing](CONTRIBUTING.md) document outlines the process to help get your
contribution accepted.

## Support & Feedback

We have a slack channel!
[Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free
to join for any support questions or feedback, we are happy to help. To report
an issue or to request a feature, open an issue
[here](https://github.com/brigadecore/brigade/issues).
