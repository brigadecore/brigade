---
title: Advanced Scripting Guide
description: 'This guide provides some tips and ideas for advanced scripting.'
aliases:
  - /advanced.md
  - /topics/advanced.md
  - /topics/scripting/advanced.md
---

# Advanced Scripting Guide

This guide provides some tips and ideas for advanced scripting. It assumes that
you are familiar with [the scripting guide](scripting.md) and the 
[JavaScript API](javascript.md).

## Using `async` and `await` to run Jobs

Recent versions of JavaScript added a new way of declaring asynchronous methods, and then calling them. This way is compatible with promises. Brigade supports the new `async` and `await` decorators.

Here's a simple [Promise chain](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise) that calls two jobs:

```javascript
const { events, Job } = require("brigadier");

events.on("exec", exec);

function exec(e, p) {
    let j1 = new Job("j1", "alpine:3.7", ["echo hello"]);
    let j2 = new Job("j2", "alpine:3.7", ["echo goodbye"]);

    j1.run()
    .then(() => {
        return j2.run()
    })
    .then(() => {
        console.log("done");
    });
};
```
[advanced-01.js](../../examples/advanced-01.js)

In the example above, we use implicit JavaScript `Promise` objects for chaining two jobs, then printing `done` after the two jobs are run. Each `Job.run()` call returns a `Promise`, and we call that `Promise`'s `then()` method.

We can rewrite this to use `await` and get the same result:

```javascript
const { events, Job } = require("brigadier");

events.on("exec", exec);

async function exec(e, p) {
    let j1 = new Job("j1", "alpine:3.7", ["echo hello"]);
    let j2 = new Job("j2", "alpine:3.7", ["echo goodbye"]);

    await j1.run();
    await j2.run();
    console.log("done");
}
```
[advanced-02.js](../../examples/advanced-02.js)

The first thing to note about this example is that we are annotating our `exec()` function with the `async` prefix. This tells the JavaScript runtime that the function is an asynchronous handler.

The two `await` statements will cause the job runs to [run synchronously](https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/await). The first job will run to completion, then the second job will run to completion. Then the `console.log` function will execute.

Note that when errors occur, they are thrown as exceptions. To handle this case, use `try`/`catch` blocks:

```javascript
const { events, Job } = require("brigadier");

events.on("exec", exec);

async function exec(e, p) {
    let j1 = new Job("j1", "alpine:3.7", ["echo hello"]);
    // This will fail
    let j2 = new Job("j2", "alpine:3.7", ["exit 1"]);

    try {
        await j1.run();
        await j2.run();
        console.log("done");
    } catch (e) {
        console.log(`Caught Exception ${e}`);
    } 
};
```
[advanced-03.js](../../examples/advanced-03.js)

In the example above, the second job (`j2`) will execute `exit 1`, which will cause the container to exit with an error. When `await j2.run()` is executed, it will throw an exception because `j2` exited with an error. In our `catch` block, we print the error message that we receive.

If we run this, we'll see something like this:

```console
$ brig run -f advanced-03.js brigadecore/empty-testbed
Event created. Waiting for worker pod named "brigade-worker-01ckcc06200kqdvkdp3nc65bap".
Build: 01ckcc06200kqdvkdp3nc65bap, Worker: brigade-worker-01ckcc06200kqdvkdp3nc65bap
prestart: no dependencies file found
prestart: src/brigade.js written
[brigade] brigade-worker version: 0.15.0
[brigade:k8s] Creating PVC named brigade-worker-01ckcc06200kqdvkdp3nc65bap
// Omitted status messages
[brigade:k8s] brigade/j2-01ckcc06200kqdvkdp3nc65bap phase Failed
  Error: Pod j2-01ckcc06200kqdvkdp3nc65bap failed to run to completion

  - k8s.js:417 k.readNamespacedPod.then.response
    ./dist/k8s.js:417:32


Caught Exception Error: job j2(j2-01ckcc06200kqdvkdp3nc65bap): Error: Pod j2-01ckcc06200kqdvkdp3nc65bap failed to run to completion

[brigade:app] after: default event handler fired
[brigade:app] beforeExit(2): destroying storage
[brigade:k8s] Destroying PVC named brigade-worker-01ckcc06200kqdvkdp3nc65bap
```

The line `Caught Exception...` shows the error that we received.

Some people feel that using `async`/`await` makes code more readable. Others prefer the `Promise` notation. Brigade will support either. The pattern above can be used with `Group` and other `Promise`-aware Brigade objects as well.

## Using Object-oriented JavaScript to Extend `Job`

JavaScript supports class-based object oriented programming. And Brigade, written in TypeScript, provides some useful ways of working with the `Job` class. The `Job` class can be extended to either preconfigure similar jobs or to add extra functionality to a job.

The following example creates a `MyJob` class that extends `Job` and provides some predefined
fields:

```javascript
const {events, Job, Group} = require("brigadier");

class MyJob extends Job {
  constructor(name) {
    super(name, "alpine:3.7");
    this.tasks = [
      "echo hello",
      "echo world"
    ];
  }
}

events.on("exec", (e, p) => {
  const j1 = new MyJob("j1")
  const j2 = new MyJob("j2")

  Group.runEach([j1, j2])
});
```
[advanced-04.js](../../examples/advanced-04.js)

In the example above, both `j1` and `j2` will have the same image and the same tasks. They inherited these predefined settings from the `MyJob` class. Using inheritence in this way can reduce boilerplate code.

The fields can be selectively overwritten, as well. So we could, for example, add another task to the first job without impacting the second job:

```javascript
const {events, Job, Group} = require("brigadier");

class MyJob extends Job {
  constructor(name) {
    super(name, "alpine:3.7");
    this.tasks = [
      "echo hello",
      "echo world"
    ];
  }
}

events.on("exec", (e, p) => {
  const j1 = new MyJob("j1")
  j1.tasks.push("echo goodbye");
  
  const j2 = new MyJob("j2")

  Group.runEach([j1, j2])
});
```
[advanced-05.js](../../examples/advanced-05.js)


If we were to look at the output of these two jobs, we'd see something like this:

```console
$ brig build logs --last --jobs
# ...
==========[  j1-01ckccs3vs14qzjma4z1zyrjas  ]==========
hello
world
goodbye

==========[  j2-01ckccs3vs14qzjma4z1zyrjas  ]==========
hello
world
```

Job `j1` has our extra command, while `j2` only inherited the defaults from `MyJob`.


## Using Docker Within a Brigade Job

It is possible to use Docker inside of a Brigade job. However, you will need to do some extra work. Because a Job must run in privileged mode to use the Docker socket, the method here presents a security risk and should not be allowed for untrusted brigade scripts.

Before you can write scripts that use privileged mode, you will need to set the following permissions on your Brigade project:

```yaml
allowPrivilegedJobs: "true"
```

To use Docker-in-Docker inside of a job, you will need to do three things:

- Select a container image for your job that can use Docker in Docker
- Set the job to `privileged = true`
- Run extra tasks to setup Docker-in-Docker (see the `dockerd-entrypoint.sh &` command below)

```javascript
  let dind = new Job("dind-run", "docker:dind");
  dind.privileged = true; // allowPrivilegedJobs must be set to true for this to work
  dind.tasks = [
      "dockerd-entrypoint.sh &", // <-- this sets up the Docker in Docker daemon
      "sleep 20", // Wait for the dockerd to start
      "echo ready to do docker builds and things."
  ];
```

Normally, you would create your own Docker image that used `FROM docker:dind` and then added your own code, but the above shows you the main steps necessary.
