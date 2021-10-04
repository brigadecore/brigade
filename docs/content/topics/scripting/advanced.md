---
title: Advanced Scripting Guide
description: This guide provides some tips and ideas for advanced scripting
section: scripting
weight: 3
aliases:
  - /advanced.md
  - /topics/advanced.md
  - /topics/scripting/advanced.md
---

# Advanced Scripting Guide

This guide provides some tips and ideas for advanced scripting. It assumes
familiarity with [the scripting guide] and the [Brigadier API].

[the scripting guide]: /topics/scripting/guide
[Brigadier API]: /topics/scripting/brigadier

## Promises and the `async` and `await` decorators

Brigade supports the various methods provided by the JavaScript language for
controlling the flow of asynchronous execution. This includes chaining together
promises as well as utilization of the `async` and `await` decorators.

Here is an example that uses a [Promise chain] to organize the execution of two
jobs:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

events.on("brigade.sh/cli", "exec", exec);

function exec(event) {
    let j1 = new Job("j1", "alpine:3.14", event);
    j1.primaryContainer.command = ["echo"];
    j1.primaryContainer.arguments = ["hello " + event.payload];

    let j2 = new Job("j2", "alpine:3.14", event);
    j2.primaryContainer.command = ["echo"];
    j2.primaryContainer.arguments = ["goodbye " + event.payload];

    j1.run()
    .then(() => {
        return j2.run()
    })
    .then(() => {
        console.log("done");
    });
}

events.process();
```

In the example above, we use implicit JavaScript `Promise` objects for chaining
two jobs, then printing `done` after the two jobs are run. Each `Job.run()`
call returns a `Promise`, and we call that `Promise`'s `then()` method.

Here's what it looks like when the script is run:

```console
$ brig event create --project promises --payload world --follow

Created event "882f832a-c156-4afc-9936-00d3b2d61083".

Waiting for event's worker to be RUNNING...
2021-10-04T22:33:22.078Z INFO: brigade-worker version: 5c94a15-dirty
2021-10-04T22:33:22.502Z [job: j1] INFO: Creating job j1
2021-10-04T22:33:25.052Z [job: j2] INFO: Creating job j2
done

$ brig event logs --id 882f832a-c156-4afc-9936-00d3b2d61083 --job j1

hello world

$ brig event logs --id 882f832a-c156-4afc-9936-00d3b2d61083 --job j2

goodbye world
```

We can rewrite the example to use [await] and get the same result:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

events.on("brigade.sh/cli", "exec", exec);

async function exec(event) {
    let j1 = new Job("j1", "alpine:3.14", event);
    j1.primaryContainer.command = ["echo"];
    j1.primaryContainer.arguments = ["hello " + event.payload];

    let j2 = new Job("j2", "alpine:3.14", event);
    j2.primaryContainer.command = ["echo"];
    j2.primaryContainer.arguments = ["goodbye " + event.payload];

    await j1.run();
    await j2.run();
    console.log("done");
}

events.process();
```

The first thing to note about this example is that we are annotating our
`exec()` function with the `async` prefix. This tells the JavaScript runtime
that the function is an asynchronous handler.

The two `await` statements will cause the jobs to [run synchronously][await].
The first job will run to completion, then the second job will run to
completion. Then the `console.log` function will execute.

Note that when errors occur, they are thrown as exceptions. To handle this
case, use `try`/`catch` blocks:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

events.on("brigade.sh/cli", "exec", exec);

async function exec(event) {
    let j1 = new Job("j1", "alpine:3.14", event);
    j1.primaryContainer.command = ["echo"];
    j1.primaryContainer.arguments = ["hello " + event.payload];

    // j2 is configured to fail
    let j2 = new Job("j2", "alpine:3.14", event);
    j2.primaryContainer.command = ["exit"];
    j2.primaryContainer.arguments = ["1"];

    try {
        await j1.run();
        await j2.run();
        console.log("done");
    } catch (e) {
        console.log(`Caught Exception ${e}`);
    }
}

events.process();
```

In the example above, the second job (`j2`) will execute `exit 1`, which will
cause the container to exit with an error. When `await j2.run()` is executed,
it will throw an exception because `j2` exited with an error. In our `catch`
block, we print the error message that we receive.

If we run this, we'll see something like this:

```console
$ brig event create --project await --payload world --follow

Created event "69b5713f-b612-434f-9b52-9bcd57f044c5".

Waiting for event's worker to be RUNNING...
2021-10-04T22:45:45.808Z INFO: brigade-worker version: 5c94a15-dirty
2021-10-04T22:45:46.235Z [job: j1] INFO: Creating job j1
2021-10-04T22:45:48.826Z [job: j2] INFO: Creating job j2
Caught Exception Error: Job "j2" failed
```

The line `Caught Exception...` shows the error that we received.

Some people feel that using `async`/`await` makes code more readable. Others
prefer the `Promise` notation. Brigade will support either. The pattern above
can also be used with `Job.concurrent()` and `Job.sequence()`, as their `run()`
methods return `Promise` objects as well.

[Promise chain]: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise
[await]: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/await

## Using Object-oriented JavaScript to Extend `Job`

JavaScript supports class-based, object-oriented programming. And Brigade,
written in TypeScript, provides some useful ways of working with the `Job`
class. The `Job` class can be extended to either preconfigure similar jobs or
to add extra functionality to a job.

The following example creates a `MyJob` class that extends `Job` and provides
some predefined fields:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

class MyJob extends Job {
  constructor(name, event) {
    super(name, "alpine:3.14", event);
    this.primaryContainer.command = ["echo"];
    this.primaryContainer.arguments = ["hello " + event.payload];
  }
}

events.on("brigade.sh/cli", "exec", async event => {
  const j1 = new MyJob("j1", event);
  const j2 = new MyJob("j2", event);

  await Job.sequence(j1, j2).run();
});

events.process();
```

In the example above, both `j1` and `j2` will have the same image and the same
command. They inherited these predefined settings from the `MyJob` class. Using
inheritence in this way can reduce boilerplate code.

The fields can be selectively overwritten, as well. We could, for example,
override the command arguments for the second job without affecting the first:

```javascript
const { events, Job } = require("@brigadecore/brigadier");

class MyJob extends Job {
  constructor(name, event) {
    super(name, "alpine:3.14", event);
    this.primaryContainer.command = ["echo"];
    this.primaryContainer.arguments = ["hello " + event.payload];
  }
}

events.on("brigade.sh/cli", "exec", async event => {
  const j1 = new MyJob("j1", event);
  const j2 = new MyJob("j2", event);
  j2.primaryContainer.arguments = ["goodbye " + event.payload];

  await Job.sequence(j1, j2).run();
});

events.process();
```

If we were to look at the output of these two jobs, we'd see something like this:

```console
$ brig event create --project jobs --payload world --follow

Created event "c4906ec3-fec1-400f-8d8f-89fd6a379475".

Waiting for event's worker to be RUNNING...
2021-10-04T23:02:58.191Z INFO: brigade-worker version: 5c94a15-dirty
2021-10-04T23:02:58.545Z [job: j1] INFO: Creating job j1
2021-10-04T23:03:01.088Z [job: j2] INFO: Creating job j2

$ brig event logs --id c4906ec3-fec1-400f-8d8f-89fd6a379475 --job j1

hello world

$ brig event logs --id c4906ec3-fec1-400f-8d8f-89fd6a379475 --job j2

goodbye world
```

Job `j2` has the different command, while `j1` inherited the defaults from `MyJob`.
