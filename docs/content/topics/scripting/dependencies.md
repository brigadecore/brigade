---
title: Dependencies
description: How dependencies work in Brigade
section: scripting
weight: 4
aliases:
  - /dependencies.md
  - /topics/dependencies.md
  - /topics/scripting/dependencies.md
---

# Import dependencies in your Brigade script

A Brigade worker is responsible for executing your Brigade script. By default,
Brigade comes with a general purpose worker which does not have any external
dependency that is not critical to executing event handlers in Brigade.

If you want to have other dependencies available in your worker execution
environment (and available in the script itself), there are multiple
approaches:

- Create a custom worker container image, which has your dependencies. This
  approach is described in detail in the [Workers] document. In a nutshell,
  use this approach if you have the same dependency for multiple projects, or
  if your dependencies take a long time to pull.

- Using the default Brigade Worker image:
  - Supply a [package.json] file containing the dependencies specific to a
    Brigade project.
  - Directly use the local dependencies located in your project's git
    repository.

This document describes the latter two approaches.

[Workers]: /topics/scripting/workers

## Add custom dependencies using a `package.json` file

If you need different dependencies for every Brigade project, this can be
easily achieved using a [package.json] file.  This can be placed alongside the
Brigade script in the project repository or embedded in the project definition.

We'll leave describing the full usage of this file to the
[official documentation][package.json], but here we'll look at the specific
case of listing dependency names and versions. These can be added under the
`dependencies` section, like so:

```json
{
    "dependencies": {
        "is-thirteen": "2.0.0"
    }
}
```

Before starting to execute the `brigade.js` script, the worker will install the  
dependencies using `npm` (or `yarn` if a `yarn.lock` file is present), adding
them to the `node_modules` folder.

Then, in the Brigade script, the new dependency can be used just like any 
other NodeJS dependency:

```javascript
const { events } = require("@brigadecore/brigadier");
const is = require("is-thirteen");

events.on("brigade.sh/cli", "exec", async event => {
  console.log("is 13 thirteen? " + is(13).thirteen());
});

events.process();
```

Now if we create an event for a project that uses this script (we've also set
`logLevel` to `DEBUG`), we will see npm being used to install dependencies, as
well as the console log that uses `is-thirteen`:

```
$ brig event create --project dependencies --follow

Created event "7987e2bb-5ca9-4f67-8d32-9f5dd667c0c5".

Waiting for event's worker to be RUNNING...
2021-09-27T22:35:23.234Z INFO: brigade-worker version: 9b52569-dirty
2021-09-27T22:35:23.239Z DEBUG: using npm as the package manager
2021-09-27T22:35:23.239Z DEBUG: found a package.json at /var/vcs/examples/13-dependencies/.brigade/package.json
2021-09-27T22:35:23.240Z DEBUG: installing dependencies using npm

added 2 packages, and audited 3 packages in 1s

found 0 vulnerabilities
2021-09-27T22:35:24.742Z DEBUG: path /var/vcs/examples/13-dependencies/.brigade/node_modules/@brigadecore does not exist; creating it
2021-09-27T22:35:24.743Z DEBUG: polyfilling @brigadecore/brigadier with /var/brigade-worker/brigadier-polyfill
2021-09-27T22:35:24.745Z DEBUG: found nothing to compile
2021-09-27T22:35:24.747Z DEBUG: running node brigade.js
is 13 thirteen? true
```

Notes:

- All custom dependencies declared in the `package.json` file will be added in
  the node process dedicated to the script environment itself, separate from
  the worker's node process and dependencies.

- Dependencies are dynamically installed on every Brigade script execution -
  this means if the dependencies added are large, and the event frequency is
  high for a particular project, it might make sense to make a pre-built Docker
  image that already contains the dependencies. See the [Workers] document for
  further details on how to do so.

[package.json]: https://docs.npmjs.com/cli/v7/configuring-npm/package-json
[Workers]: /topics/scripting/workers

## Using local dependencies from the project repository

Local dependencies are resolved using standard Node [module resolution]. This
approach works great for using dependencies that are not intended to be
external packages, and which are located in the project repository.

These dependencies may be placed in the default configuration directory for
Brigade, `./brigade`, alongside other config files like the project script
(e.g. `brigade.js`) and `package.json`.

Let's consider the following scenario: we have a JavaScript file located in
`/.brigade/circle.js`. In our Brigade script, we can use any exported method or
variable from that package by simply using a `require` statement, just like in
any other JavaScript project.

```javascript
// file /.brigade/circle.js
var PI = 3.14;
exports.area = function (r) {
    return PI * r * r;
};
exports.circumference = function (r) {
    return 2 * PI * r;
};
```

Then, in our `brigade.js` we can import that file and use it:

```javascript
const { events } = require("@brigadecore/brigadier");
const circle = require("./circle");

events.on("brigade.sh/cli", "exec", async event => {
  console.log("area of a circle with radius 3: " + circle.area(3));
});

events.process();
```

Here is the output when we create an event via `brig` for a project using this
script (plus `logLevel: DEBUG`):

```console
$ brig event create --project dependencies --follow

Created event "8aa3c5dd-a685-493a-a366-a6183a9e2650".

Waiting for event's worker to be RUNNING...
2021-09-28T13:43:49.143Z INFO: brigade-worker version: 9b52569-dirty
2021-09-28T13:43:49.148Z DEBUG: using npm as the package manager
2021-09-28T13:43:49.148Z DEBUG: path /var/vcs/examples/13-dependencies/.brigade/node_modules/@brigadecore does not exist; creating it
2021-09-28T13:43:49.149Z DEBUG: polyfilling @brigadecore/brigadier with /var/brigade-worker/brigadier-polyfill
2021-09-28T13:43:49.149Z DEBUG: found nothing to compile
2021-09-28T13:43:49.150Z DEBUG: running node brigade.js
area of a circle with radius 3: 28.259999999999998
```

[module resolution]: https://nodejs.org/api/modules.html#modules_all_together

## Both approaches in one example

Check out the [13-dependencies] example project to see both approaches
incorporated into one project. Feel free to create the project, create events
for the project, etc., to get a feel for how both methods work.

[13-dependencies]: https://github.com/brigadecore/brigade/tree/main/examples/13-dependencies
