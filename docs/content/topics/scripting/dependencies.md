---
title: Dependencies
description: How dependencies work in Brigade
aliases:
  - /dependencies.md
  - /topics/dependencies.md
  - /topics/scripting/dependencies.md
---

TODO: update per v2

# Import dependencies in your `brigade.js` file

A Brigade worker is responsible for executing your `brigade.js` file. By default, Brigade comes with a general purpose worker which does not have any external dependency that is not critical to controlling the flow of your pipeline.

If you want to have other dependencies available in your worker execution environment (and available in `brigade.js`), there are multiple approaches:

- by creating a custom worker container image, which has your dependencies. This approach is [described in detail in this document](workers.md). In a nutshell, use this approach if you have the same dependency for multiple projects, or if your dependencies take a long time to pull.

- without creating a custom container image:
    - by supplying a `brigade.json` file (similar to a `package.json` file) 
that contains the dependencies and that are specific on every Brigade project.
    - by directly using local dependencies located in your project repository.

This document describes the last two approaches.

> Here you can find a [repository that exemplifies both approaches here](https://github.com/radu-matei/brigade-javascript-deps).

## Add custom dependencies using a `brigade.json` file

If you need different dependencies for every Brigade project, this can be easily achieved 
using a `brigade.json` file.  This can be placed side-by-side the `brigade.js` file in the project
repository and/or supplied at runtime via a `brig run` command.  See the [brig](brig.md) doc for
details on the latter method.  Note that if a `brigade.json` is supplied at runtime, this takes
precedence over the file found in version control.

This file is intended to hold general configuration details for Brigade.  The list of dependency
names and versions can be added under the `dependencies` section, like so:

```
{
    "dependencies": {
        "is-thirteen": "2.0.0"
    }
}
```
Before starting to execute the `brigade.js` script, the worker will install the  
dependencies using `yarn`, adding them to the `node_modules` folder.

Then, in the `brigade.js` file, the new dependency can be used just like any 
other NodeJS dependency:

```
const { events } = require("brigadier")
const is = require("is-thirteen");

events.on("exec", function (e, p) {
    console.log("is 13 thirteen? " + is(13).thirteen());
})
```

Now if we run a build for this project, we see the `is-thirteen` dependency added, 
as well as the console log resulted from using it:

```
$ brig run brigade-86959b08a89af5b6b83f0ace6d9030f1fdca7ed8ea0a296e27d72e
Event created. Waiting for worker pod named "brigade-worker-01cexrvrs08shcev26961cwd6n".
Started build 01cexrvrs08shcev26961cwd6n as "brigade-worker-01cexrvrs08shcev26961cwd6n"
installing is-thirteen@2.0.0
prestart: src/brigade.js written
[brigade] brigade-worker version: 0.14.0
[brigade:k8s] Creating PVC named brigade-worker-01cexrvrs08shcev26961cwd6n
is 13 thirteen? true
[brigade:app] after: default event handler fired
[brigade:app] beforeExit(2): destroying storage
[brigade:k8s] Destroying PVC named brigade-worker-01cexrvrs08shcev26961cwd6n
```

Notes:

- when adding a custom dependency using `brigade.json`, `yarn` will add it side-by-side with [the worker's 
dependencies](../../brigade-worker/package.json) - this means that the process will fail if a dependency that conflicts with one of the 
worker's dependencies is added. However, already existing dependencies (such as `@kubernetes/client-node`, `ulid` or `chai`) 
can be used from `brigade.js` without adding them to `brigade.json`. 

However, the only issue is trying to add a different version of an already existing dependency of the worker.

- as the Brigade worker is capable of modifying the state of the cluster, be mindful 
of any external dependencies added through `brigade.json`

- for now, the only part of `brigade.json` used is the `dependencies` object - it means the rest of the file 
can be used to pass additional information to `brigade.js`, such as environment data - however, as the `brigade.json` 
file is passed in the source control repository, information passed there will be available to anyone with access to the repository.

- all dependencies added in `brigade.json` are dynamically installed on every Brigade build - this means if the dependencies added 
are large, and the build frequency is high for a particular project, it might make sense to make a pre-built Docker image that 
already contains the dependencies, [as described in this document](workers.md).

## Using local dependencies from the project repository

Local dependencies are resolved using standard Node [module resolution](https://nodejs.org/api/modules.html#modules_all_together),
with one change: the worker's `node_modules` directory is added as a fallback, so `brigade.js`—and any local dependencies—can resolve modules installed via `brigade.json`.
This approach works great for using dependencies that are not intended to be external packages, and which are located in the project repository. 

Let's consider the following scenario: we have a JavaScript file located in `/local-deps/circle.js`, where `local-deps` is a directory at the root of our git repository. In our `brigade.js` file, we can use any exported method or variable from that package by simply using a `require` statement, just like in any other JavaScript project.

```javascript
// file /local-deps/circle.js
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
const { events } = require("@brigadecore/brigadier")
const circle = require("./local-deps/circle");

events.on("exec", function (e, p) {
    console.log("area of a circle with radius 3 " + circle.area(3));
});
```

## Best Practices

As we have seen, it is easy to add new functionality to the Brigade worker. But
it is important to keep in mind that the Worker is intended to do one thing:
execute Brigade chains.

To that end, it is best to fight the temptation to put too much logic into the
Brigade worker. Where possible, use Jobs to perform specific tasks within their
own containers, and use workers to control the execution of a series of Jobs.