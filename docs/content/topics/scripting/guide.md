---
title: Scripting Guide
description: 'How to create and structure `brigade.js` files.'
aliases:
  - /guide.md
  - /topics/guide.md
  - /topics/scripting/guide.md
---

TODO: update per v2

# Scripting Guide

This guide explains the basics of how to create and structure `brigade.js` files.

# Brigade Scripts, Projects, and Repositories

Brigade scripts are stored in a `brigade.js` file. They are designed to contain short scripts
that coordinate running multiple related tasks. We like to think of these as
cluster-oriented shell scripts: The script just ties together a bunch of other programs.

At the very core, a Brigade script is simply a JavaScript file that is executed within a
special cluster environment. An environment is composed of the following things:

- A Brigade server running in your cluster
- A [project](./projects.md) in which the script will run
- (Optionally) A source code repository (e.g. git) that contains data our
- (Optionally) A [`brigade.json`](workers.md) file that contains additional dependencies that can be used from `brigade.js`
- (Optionally) Other configuration data, like [secrets](secrets.md)

For the examples in this document, we have created a project via `brig` with these values:

```console
$ brig project create
? VCS or no-VCS project? VCS
? Project Name brigadecore/empty-testbed
? Full repository name github.com/brigadecore/empty-testbed
? Clone URL (https://github.com/your/repo.git) https://github.com/brigadecore/empty-testbed.git
? Add secrets? Yes
? 	Secret 1 dbPassword
? 	Value supersecret
? ===> Add another? No
Auto-generated a Shared Secret: "QwSBinN9sHZyuSYyTKmVOIAk"
? Configure GitHub Access? No
? Configure advanced options No
Project ID: brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
```

Note that we have linked this project to a GitHub repository. That repository
contains a very simple Node.js application. (Of course, Brigade itself does not
care what a repository contains.)

Based on the above, we have:

- A project named `brigadecore/empty-testbed`
- A GitHub repo at `https://github.com/brigadecore/empty-testbed.git`
- And a single secret: `dbPassword: supersecret`

We'll use this project throughout the tutorial as we create some `brigade.js` files and test
them out.

We will be using the `brig` binary to execute our brigade scripts.  Usage
and installation instructions for `brig` are [here](/brig/).

## A Basic Brigade Script

Here is a basic `brigade.js` file that just sends a single
message to the log:

```javascript
console.log("hello world")
```
[brigade-01.js](../../examples/brigade-01.js)

If we run this script, we would see output that looked something like this:

```console
Started brigade-worker-01brwzs64rve2jvky87hxy1wsp-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
hello world
Done in 1.44s.
```

> Tip: You can use the `brig`  command to send brigade.js
> files to Brigade.
>
> In this tutorial we run scripts with `brig run --file brigade.js $PROJECT`, where `$PROJECT` is a
> project ID like `brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac`.

In essence, all we have done is started a brigade session, logged "hello world", and exited.

This example executes the log in the _global scope_. While it's fine to work in the
global scope, most of the time what we want to do with Brigade is write _event handlers_.

## Brigade Events and Event Handlers

Brigade listens for certain things to happen, such as a new push on a GitHub repository or an
image update on DockerHub. (The events that it listens for are configured in your
project).

When Brigade observes such an event, it will load the `brigade.js` file and see if there
is an _event handler_ that matches the _event_.

Using the `brig` tool introduced above, we can see how this works.

The `brig` tool triggers an `exec` event (short for _execute_) on Brigade. So our
`brigade.js` file can intercept this event and respond to it:

```javascript
const { events } = require("brigadier")

events.on("exec", () => {
  console.log("==> handling an 'exec' event")
})
```
[brigade-02.js](../../examples/brigade-02.js)

> The `() => {}` syntax is a newer way to declare an anonymous function. We use this
> syntax throughout the tutorial, but unless otherwise noted, you can substitute in the
> traditional syntax: `function () {}`.

There are a few things to note about this script:

- It imports the `events` object from `brigadier`. Almost all Brigade scripts do this.
- It declares a single event handler that says "on an 'exec' event, run this function".

Similarly to our first script, this function simply dumps a message to a log.

The above produces something like this:

```console
Started brigade-worker-01brx1z21af78hsj4q55anycpc-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01brx1z21af78hsj4q55anycpc-master
==> handling an 'exec' event
Destroying PVC named brigade-worker-01brx1z21af78hsj4q55anycpc-master
Done in 1.49s.
```

One of the strengths of Brigade is that we can handle different events in the same file.
Brigade will only trigger the appropriate events. For example, we can expand the example
above to also provide a handler for the GitHub `push` event:

```javascript
const { events } = require("brigadier")

events.on("exec", () => {
  console.log("==> handling an 'exec' event")
})

events.on("push", () => {
  console.log(" **** I'm a GitHub 'push' handler")
})
```
[brigade-03.js](../../examples/brigade-03.js)

Now if we re-run our `brig` command, we will see the same output as before:

```
Started brigade-worker-01brx5m1yppb0dxn4emk76jqtv-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01brx5m1yppb0dxn4emk76jqtv-master
==> handling an 'exec' event
Destroying PVC named brigade-worker-01brx5m1yppb0dxn4emk76jqtv-master
Done in 1.21s.
```

Since Brigade did not see a `push` event, it did not execute the `push` event handler.
It only executed the `exec` handler that `brig` causes.

### Where Do Events Come From?

In order to be able to write good Brigade scripts, we need to know what events we
can listen for. So what is the origin of these events?

To answer this question, we need a tiny bit of background knowledge about Brigade. Brigade has
several components in its runtime. The _worker_ runs the JavaScript that we are writing
here. The _controller_ starts and manages the _worker_. But there are also Brigade
_gateways_.

A Brigade gateway is a piece of code that executes in your cluster and generates events.  An
overview of a few Brigade gateways ready for use can be seen in the [Gateways doc](./gateways.md).

The `brig` client declares its own hook (`exec`).

You can create your own gateways which generate their own events, or use gateways created
by others.

In the list above, there are two special events that have very specific usage: `after` and
`error`.

### Two Special Events: _after_ and _error_

The `after` and `error` events are what are called _lifecycle hooks_. They let you
execute some code when Brigade hits a certain stage of processing.

The `after` hook runs _at the very end_ of any session that did not die from an error.

```javascript
const { events } = require("brigadier")

events.on("exec", () => {
  console.log("==> handling an 'exec' event")
})

events.on("after", () => {
  console.log(" **** AFTER EVENT called")
})
```
[brigade-04.js](../../examples/brigade-04.js)

The `brig` client will trigger the `exec` event. But then when that event has been
processed, Brigade will trigger an `after` event before returning:

```console
Started brigade-worker-01brx76gx5v3e8r6vbmzcda7q9-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01brx76gx5v3e8r6vbmzcda7q9-master
==> handling an 'exec' event
 **** AFTER EVENT called
Destroying PVC named brigade-worker-01brx76gx5v3e8r6vbmzcda7q9-master
Done in 1.19s.
```

Note that the `**** AFTER EVENT called` message is printed right before the shutdown
begins.

The `error` event is similar, but it is _only_ triggered when a script encounters an
error.

The `after` and `error` events give you an opportunity to do some final processing (like
maybe sending a message to your chat server) before exiting.

At this point, we've taken a look at Brigade's event system. But we haven't really done
anything with Brigade except log a few messages. Let's turn out attention toward Brigade's
ability to "script up" containers.

## Jobs and Containers

A typical Brigade script divides up the workload like this:

- An _event handler_ declares how an event (`push`, `exec`) is processed
- When Brigade triggers an event, it creates a new `build`, which you can think of as "a
  specific run of your `brigade.js` file.
- A build has several _jobs_, where each job starts a container
- A job runs zero or more _tasks_ inside of a container

And in the next section, we'll talk about how we can _group_ jobs.

In the last section, we focused on the event handlers, and we ran several builds that just
logged messages. In this section, we'll create a few jobs.

To start with, let's create a simple job that doesn't really do any work:

```javascript
const { events, Job } = require("brigadier")

events.on("exec", () => {
  var job = new Job("do-nothing", "alpine:3.4")

  job.run()
})
```
[brigade-05.js](../../examples/brigade-05.js)

The first thing to note is that _we have changed the first line_. In addition to importing
the `events` object, we are now also importing `Job`. `Job` is the prototype for creating
all jobs in Brigade.

Next, inside of our `exec` event handler, we create a new `Job`, and we give it two pieces
of information:

- a _name_: The name must be unique to the event handler and should be composed of
  letters, numbers, and dashes (`-`).
- an _image_: This can be _any image that your cluster can find_. In the case above, we
  use the image named `alpine:3.4`, which is fetched [straight from DockerHub](https://hub.docker.com/_/alpine/).

The image is a crucial part of the Brigade ecosystem. A container is created from this
image, and it's within this container that we do the "real work". At the beginning, we
explained that we think of Brigade scripts as "shell scripts for your cluster." When you
execute a shell script, it is typically some glue code that manages calling one or more
external programs in a specific way.

```
#!/bin/bash

ps -ef "hello" | grep chrome
```

The script above really just organizes the way we call two existing programs (`ps` and
`grep`). Brigade does the same, except instead of executing _programs_, it executes
_containers_. Each container is expressed as a Job, which is a wrapper that knows how to
execute containers.

So in our example above, we create a Job named "do-nothing" that runs an Alpine Linux
container and (as the name implies) does nothing.

Jobs are created and run in different steps. This way, we can do some preparation on our
job (as we will see in a moment) before executing it.

To run a job, we use the Job's `run()` method. Behind the scenes, Brigade will start a new
`alpine:3.4` pod (downloading it if necessary), execute it, and monitor it. When the
container is done executing, the run is complete.

It is worth noting that a `run()` method is _asynchronous_. That means that when we call
`run()`, it will start processing in the background. Later on, we will see how we can take
advantage of that to create pipelines.

If we run the code above, we'll get output that looks something like this:

```console
Started brigade-worker-01brx7v6wsg31k81x0h4pznv47-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01brx7v6wsg31k81x0h4pznv47-master
looking up default/github-com-brigadecore-empty-testbed-do-nothing
Creating secret do-nothing-1504055331341-master
Creating Job Cache PVC github-com-brigadecore-empty-testbed-do-nothing
Creating pod do-nothing-1504055331341-master
Timeout set at 900000
default/do-nothing-1504055331341-master phase Pending
default/do-nothing-1504055331341-master phase Pending
default/do-nothing-1504055331341-master phase Succeeded
Destroying PVC named brigade-worker-01brx7v6wsg31k81x0h4pznv47-master
Done in 5.12s.
```

That's quite a lot of new output for a program that "does nothing". The important part is
this: `Creating pod do-nothing-1504055331341-master`. That tells us that it has taken our
Job and packaged it up as what Kubernetes calls a _pod_. That means it has been scheduled
for execution.

For a few lines we will see messages that let us know that our job isn't running yet, but
is in state `Pending`. Then it will run, and if it runs to a successful completion, it
will be marked as `Succeeded`.

After that, our job is cleaned up, and so is the build.

> Tip: You can see the actual output of each Job through the Brigade user interface, or using
> the Kubernetes `kubectl` client: `kubectl logs do-nothing-1504055331341-master`

Basically, our simple build just created an empty Alpine Linux pod which had nothing to do
and so exited immediately.

### Adding Tasks to Jobs

To make our Job do more, we can add tasks to it. A task is an _individual step to be run
inside of the Job's container_.

```javascript
const { events, Job } = require("brigadier")

events.on("exec", () => {
  var job = new Job("do-nothing", "alpine:3.4")
  job.tasks = [
    "echo Hello",
    "echo World"
  ]

  job.run()
})
```
[brigade-06.js](../../examples/brigade-06.js)

Tasks can be an arbitrary shell command _that is supported by the image you use_.
Tasks are concatenated together and executed as one shell script. (In Linux, they
are executed as `/bin/sh` commands, with `set -eo pipefail`.)

> You can change the shell a job uses by setting `Job.shell`. However, if the shell
> you set is not present in the image, this will cause an error.

In the example above we have added some tasks by adding them to the tasks array: `job.tasks = [ /* ... */]`
It will run the `echo` command twice. If we run this new script, its output will look just
about the same as when we ran no tasks:

```console
Started brigade-worker-01brx98hq5f3e93jxy5ddpfwgx-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01brx98hq5f3e93jxy5ddpfwgx-master
looking up default/github-com-brigadecore-empty-testbed-do-nothing
Creating secret do-nothing-1504056818776-master
Creating Job Cache PVC github-com-brigadecore-empty-testbed-do-nothing
Creating pod do-nothing-1504056818776-master
Timeout set at 900000
default/do-nothing-1504056818776-master phase Pending
default/do-nothing-1504056818776-master phase Succeeded
Destroying PVC named brigade-worker-01brx98hq5f3e93jxy5ddpfwgx-master
Done in 5.14s.
```

But this time, we can take a look at the logs for our pod and see the results of our
tasks. We will do this using the `kubectl` tool, though you can also use the Brigade UI.

```console
$ kubectl logs do-nothing-1504056818776-master
Hello
World
```

Here's what happened: Our `brigade.js` script created a new pod named `do-nothing-1504056818776-master`, which
started `alpine:3.4` and then ran the two `echo` tasks. When it completed, Brigade let us
know that it `Succeeded` and then it finished up the build.

Now we can take things one more step and create _two jobs_ that each do something.

```javascript
const { events, Job } = require("brigadier")

events.on("exec", () => {
  var hello = new Job("hello", "alpine:3.4")
  hello.tasks = [
    "echo Hello",
    "echo World"
  ]

  var goodbye = new Job("goodbye", "alpine:3.4")
  goodbye.tasks = [
    "echo Goodbye",
    "echo World"
  ]

  hello.run()
  goodbye.run()
})
```
[brigade-07.js](../../examples/brigade-07.js)

In this example we create two jobs (`hello` and `goodbye`). Each starts an Alpine
container and prints a couple of messages, then exits.

After defining each one, we run them like this:

```javascript
hello.run()
goodbye.run()
```

Now the output of running this command with `brig` might be a little surprising:

```
Started brigade-worker-01brx9n20bsjxeweggtzb7fpka-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01brx9n20bsjxeweggtzb7fpka-master
looking up default/github-com-brigadecore-empty-testbed-hello
looking up default/github-com-brigadecore-empty-testbed-goodbye
Creating secret hello-1504057229136-master
Creating secret goodbye-1504057229149-master
Creating Job Cache PVC github-com-brigadecore-empty-testbed-hello
undefined
Creating Job Cache PVC github-com-brigadecore-empty-testbed-goodbye
undefined
Creating pod hello-1504057229136-master
Creating pod goodbye-1504057229149-master
Timeout set at 900000
Timeout set at 900000
default/hello-1504057229136-master phase Pending
default/goodbye-1504057229149-master phase Pending
default/goodbye-1504057229149-master phase Pending
default/hello-1504057229136-master phase Pending
default/goodbye-1504057229149-master phase Pending
default/hello-1504057229136-master phase Pending
default/goodbye-1504057229149-master phase Pending
default/hello-1504057229136-master phase Pending
default/goodbye-1504057229149-master phase Pending
default/hello-1504057229136-master phase Pending
default/hello-1504057229136-master phase Pending
default/goodbye-1504057229149-master phase Pending
default/hello-1504057229136-master phase Succeeded
default/goodbye-1504057229149-master phase Succeeded
Destroying PVC named brigade-worker-01brx9n20bsjxeweggtzb7fpka-master
Done in 15.17s.
```

What is surprising is that if we look at the output above, we see that both jobs seem to
be running at the same time. This is because when we start a job in Brigade, it runs
asynchronously. Another way to phrase that is that jobs run _parallel by default_.

Again, if you want to view the output of each job, you can use the `kubectl logs` command
for each pod.

If we want these two pods to run _in order_, with `hello` running to completion before
`goodbye` starts, we can do that by creating what in JavaScript is called a "Promise
chain":

```javascript
const { events, Job } = require("brigadier")

events.on("exec", () => {
  var hello = new Job("hello", "alpine:3.4")
  hello.tasks = [
    "echo Hello",
    "echo World"
  ]

  var goodbye = new Job("goodbye", "alpine:3.4")
  goodbye.tasks = [
    "echo Goodbye",
    "echo World"
  ]

  hello.run().then(() => {
    goodbye.run()
  })
})
```
[brigade-08.js](../../examples/brigade-08.js)

The important new part is at the end. We have replaced this:

```javascript
hello.run()
goodbye.run()
```

The new version looks like this:

```javascript
hello.run().then(() => {
  goodbye.run()
})
```

And we can read it like this: "run hello, then run goodbye". In the Groups section below,
we will see a simpler way of doing this. But for now, this is one way of running jobs in
sequence.

Now if we run our `brigade.js`, the output will look like this:

```console
Started brigade-worker-01brxgx62zey1b31ae2ccd2xnm-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01brxgx62zey1b31ae2ccd2xnm-master
looking up default/github-com-brigadecore-empty-testbed-hello
Creating secret hello-1504059196321-master
Creating Job Cache PVC github-com-brigadecore-empty-testbed-hello
Creating pod hello-1504059196321-master
Timeout set at 900000
default/hello-1504059196321-master phase Pending
default/hello-1504059196321-master phase Succeeded
looking up default/github-com-brigadecore-empty-testbed-goodbye
Creating secret goodbye-1504059200407-master
Creating Job Cache PVC github-com-brigadecore-empty-testbed-goodbye
Creating pod goodbye-1504059200407-master
Timeout set at 900000
default/goodbye-1504059200407-master phase Pending
default/goodbye-1504059200407-master phase Succeeded
Destroying PVC named brigade-worker-01brxgx62zey1b31ae2ccd2xnm-master
Done in 9.17s.
```

Compared to our previous version, we can see the different. It runs the `hello` job first,
and then runs the `goodbye` job.

Before moving on to talk about groups, though, let's do one short example that does
something useful. Our project points to the [empty testbed](https://github.com/brigadecore/empty-testbed)
repository in GitHub. That repository happens to have a small Node.js application, and we
can write a simple set of tasks to build and run that application.

### Running Tasks Against a Git Repository

Earlier we talked about how a project may have an associated Git repository. And when we
created our project, we pointed it to [a
repository](https://github.com/brigadecore/empty-testbed) that contains a simple Node.js
application. In this example, we are going to work directly with that repository.

Here's our new script:

```javascript
const { events, Job } = require("brigadier")

events.on("exec", () => {
  var test = new Job("test-app", "node:8")

  test.tasks = [
    "cd /src/hello",
    "yarn install",
    "node index.js"
  ]

  test.run()
})
```
[brigade-09.js](../../examples/brigade-09.js)

This time around, we are going to run three tasks:

- `cd /src/hello`: Change directories into the place where our source code is.
- `yarn install`: Install the dependencies for our Node.js app. (Yarn is like pip, Maven,
  Glide or CPAN, but for Node.js.)
- `node index.js`: Run the `index.js` file inside of `/src/hello`.

If we run the script, we'll see the usual output, and it will contain a line like this:

```
Creating pod test-app-1504064455281-master
```

If we use `kubectl` to check out the log for our `test-app` container, we'll see something
like this:

```console
$ kubectl logs test-app-1504064455281-master
yarn install v0.27.5
info No lockfile found.
[1/4] Resolving packages...
[2/4] Fetching packages...
[3/4] Linking dependencies...
[4/4] Building fresh packages...
success Saved lockfile.
Done in 0.14s.
hello world
```

That is the output of our three tasks.

But what exactly is this build doing? Why are we doing a `cd /src/hello`? And where did
that directory even come from?

Here's what is happening: Because our project has a Git repository associated with it,
Brigade is automatically getting us a copy of that project and attaching that copy to each
`Job` that we run.

When a repository is checked out, it is stored by default in `/src`. So it is as if we
started every job by doing a `git clone https://github.com/brigadecore/empty-testbed.git /src`.
That means we can start our job knowing that we already have access to everything in our
Git project.

So when we `cd` into `/src/hello`, we're changing into the `hello/` directory in the Git
project. And from there on, we are working with the code from the repository.

Being able to associated a Git repository to a project is a convenient way to provide
version-controlled data to our Brigade builds. And it makes Brigade a great tool for executing
CI pipelines, deployments, packaging tasks, end-to-end tests, and other DevOps tasks.

Automatically mounting a repository is typically a great feature. But every once in a
while it is useful to _disable_ this behavior. To do that, simply add an attribute to your
job:

```javascript
var job = new Job("test", "node:8")
job.useSource = false
```

You can also change the path where the source is stored. The default is `src/`, but it can
be set to another location:

```javascript
var job = new Job("test", "node:8")
job.mountPath = "/mnt/brigade/src"
```

## Groups

Earlier we saw an example of creating and running multiple jobs. We saw a simple form
where we could run two jobs in parallel:

```javascript
job1.run()
job2.run()
```

And we saw a slightly more complicated form where we ran one job and then another:

```javascript
job1.run().then( () => job2.run() )
```

These are two ways to work with individual jobs. But sometimes it is desirable to work
with jobs as if they were a group.

For example, we can run two jobs as an ordered group:

```javascript
const { events, Job, Group } = require("brigadier")

events.on("exec", () => {
  var hello = new Job("hello", "alpine:3.4", ["echo hello"])
  var goodbye = new Job("goodbye", "alpine:3.4", ["echo goodbye"])

  Group.runEach([hello, goodbye])
})
```
[brigade-10.js](../../examples/brigade-10.js)

There are three things to notice in the example above:

1. We now also import `Group` along with `events` and `Job`.
2. Since we are running a simple list of tasks, we declare the task list in the `Job()`
   constructor.
3. We use `Group.runEach()` to run our tasks.

> Tip: Using `Group.runEach()` is often easier to read than creating a Promise chain.

Group has a couple of useful static methods:

- `Group.runEach()` takes a list of tasks and runs them _in sequence_. It does not mark itself
  as complete until every task has been executed.
- `Group.runAll()` takes a list of tasks and runs them all _in parallel_. It does not mark
  itself complete until all of the tasks have finished.

Both of these methods return Promise objects, so they can be chained. For example, here is
an example that runs a `hello` and `goodbye` jobs _in parallel_, then it runs a
`hello-again` job only after both of the others have completed.

```javascript
const { events, Job, Group } = require("brigadier")

events.on("exec", () => {
  var hello = new Job("hello", "alpine:3.4", ["echo hello"])
  var goodbye = new Job("goodbye", "alpine:3.4", ["echo goodbye"])
  var helloAgain = new Job("hello-again", "alpine:3.4", ["echo hello again"])

  Group.runAll([hello, goodbye])
    .then( ()=> {
      helloAgain.run()
    })
})
```
[brigade-11.js](../../examples/brigade-11.js)

In the above case, `hello` and `goodbye` will run at the same time. But `helloAgain` will
not be started until both of the others have finished.

Using groups, you can create sophisticated pipelines.

Sometimes you may want to declare groups ahead of time and then run them, much as you do
with jobs. This is an alternative to using the Group static methods.

```javascript
const { events, Job, Group } = require("brigadier")

events.on("exec", () => {
  var hello = new Job("hello", "alpine:3.4", ["echo hello"])
  var goodbye = new Job("goodbye", "alpine:3.4", ["echo goodbye"])

  var helloAgain = new Job("hello-again", "alpine:3.4", ["echo hello again"])
  var goodbyeAgain = new Job("bye-again", "alpine:3.4", ["echo bye again"])


  var first = new Group()
  first.add(hello)
  first.add(goodbye)

  var second = new Group()
  second.add(helloAgain)
  second.add(goodbyeAgain)

  first.runAll().then( () => second.runAll() )
})
```
[brigade-12.js](../../examples/brigade-12.js)

The above creates two groups, and then later executes them. The order of execution would
be:

- Run both of the jobs in the `first` group.
- Once those two jobs have both completed, run both of the jobs in the `second` group.

This is the way groups can be used to control the order in which groups of jobs are run.

So far we have looked at the brigade.js file, the event registry, and jobs and groups. As we
advance our way through Brigade, we will next take a look at how `brigade.js` scripts can work
with the BrigadeEvent and Project objects that are sent to every event handler.

## Working with Event and Project Data

Two pieces of information are passed into every event handler: The event that occurred,
and the project.

### The Brigade Event

From the event, we can find out what triggered the event, what data was sent with the
event, and (if there is a repository) what repository commit we should be using.

An event looks like this:

```javascriot
var e = {
  buildID: "brigade-worker-01brwzs64rve2jvky87hxy1wsp-master",
  type: "brig",
  provider: "brig",
  commit: "master",
  payload: ""
}
```

- `buildID` is a unique per-build ID. Every time a new build is triggered, a new ID will
  be generated.
- `type` is the event type. A GitHub Pull Request, for example, would set type to
  `pull_request`.
- `provider` tells what service triggered this event. A GitHub request will set this to
  `github`.
- `commit` is the commit (revision ID) for the VCS. For Git, this can be a SHA, a branch
  name, or a tag.
- `payload` contains any information that the hook sent when triggering the event. For
  example, a GitHub push request generates [a rather large payload](https://developer.github.com/v3/activity/events/types/#pushevent). Payload is an unparsed string.

When one event triggers another event, the triggered event also gets information about the
thing that caused it. This is available on the `event.cause` field. The `after` and
`error` events, for example, will be able to access the `cause` field and see the event
that triggered them.

### The Project

The project gives us information about the repository, the Kubernetes configuration, and
secrets (environment variables or credentials) that are available to us.

```
{
  "id":"brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac",
  "name":"github.com/brigadecore/empty-testbed",
  "kubernetes":{
    "namespace":"default",
    "vcsSidecar":"brigadecore/git-sidecar:latest",
    "buildStorageSize": "50Mi"
  },
  "repo":{
    "name":"brigadecore/empty-testbed",
    "cloneURL":"https://github.com/brigadecore/empty-testbed.git"
  },
  "secrets":{
    "dbPassword":"supersecret"
  }
}
```

- `id` is the project ID, as generated by BRIGADE.
- `name` is the project name that is set when you created the project with `brig`.
- `kubernetes` stores Kubernetes-related fields:
  - `namespace` is the namespace in which Brigade runs
  - `vcsSidecar` is the container image that Brigade uses internally to check out your VCS
    repository.
  - `buildStorageSize` is the size of the build shared storage space used by the jobs.
- `repo` stores information about your VCS repository.
  - `name` is the name of the repo. GitHub projects are named as org/project.
  - `cloneURL` is the URL Brigade will use to clone or fetch the repository.
- `secrets` is where you can store environment variables or secrets. These are set during
  project creation via `brig` (see the [Secrets Guide](secrets.md).

### Using Event and Project Objects

Both the event and the project are passed to every event handler. Until now, we have
ignored them. But we can write a simple script to show some information about the event
and the project:

```javascript
const { events } = require("brigadier")

events.on("exec", (e, p) => {
  console.log(">>> event " + e.type + " caused by " + e.provider)
  console.log(">>> project " + p.name + " clones the repo at " + p.repo.cloneURL)
})
```
[brigade-13.js](../../examples/brigade-13.js)

If we run the above, we'll see output like this:

```console
Started brigade-worker-01brz271ma5h06na0bb5j7d2rm-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01brz271ma5h06na0bb5j7d2rm-master
>>> event exec caused by brig
>>> project github.com/brigadecore/empty-testbed clones the repo at https://github.com/brigadecore/empty-testbed.git
Destroying PVC named brigade-worker-01brz271ma5h06na0bb5j7d2rm-master
Done in 1.04s.
```

Event and project data should be treated with a little extra care. Things like
`secrets` or event `cloneURL` might not be the sorts of information you want accidentally
displayed.

### Passing Project or Event Data to Jobs

Brigade is designed to make it easy for you to extract information from the event and project
and sent it into a job. Here are two ways to share information with jobs:

```javascript
const { events, Job } = require("brigadier")

events.on("exec", (e, p) => {
  var echo = new Job("echo", "alpine:3.4")
  echo.tasks = [
    "echo Project " + p.name,
    "echo Event $EVENT_NAME"
  ]

  echo.env = {
    "EVENT_NAME": e.type
  }

  echo.run()
})
```
[brigade-14.js](../../examples/brigade-14.js)

In the above code, we create a job named `echo` and we run two tasks. In the first, we
directly inject the project name (`p.name`) into the task command before the task is run.

In the second case, we use environment variables to pass a name/value pair into the
command, and then it is evaluated at runtime.

`echo.env` is the place to set environment variables for the container. The variable set
to `EVENT_NAME` there is accessible inside the pod as `$EVENT_NAME`.

If we look at the output of the pod, we'll see this:

```
$ kubectl logs echo-1504074306432-master
Project github.com/brigadecore/empty-testbed
Event exec
```

At this point we have seen how we can access information about the project and event. In
the next section we are going to turn to storage. We are going to see how Brigade provides
ways for builds and jobs to share storage space.

## Storing Data with Caches and Shared Space

There are many ways that developers can store and retrieve data from Brigade. For example,
object storage systems like Azure Object Storate or Amazon S3 or hosted database
providers. Brigade developers may choose to use those tools from within jobs.

But Brigade comes with a few built-in options that are useful in writing basic Brigade builds,
and which don't require modifying your containers. In this part of the guide, we cover two
built-in shared directories.

### Build Storage (Shared Space)

Each build gets its own shared storage. This is a file path that can be accessed by every
job during the build, but which does not survive after the build has completed.

When enabled, storage is mounted to `/mnt/brigade/share` on each job's container.

```javascript
const { events, Job, Group } = require("brigadier")

events.on("exec", (e, p) => {
  var dest = "/mnt/brigade/share/hello.txt"
  var one = new Job("one", "alpine:3.4", ["echo hello > " + dest])
  var two = new Job("two", "alpine:3.4", ["echo world >> " + dest])
  var three = new Job("three", "alpine:3.4", ["cat " + dest])

  one.storage.enabled = true
  two.storage.enabled = true
  three.storage.enabled = true

  Group.runEach([one, two, three])
})
```
[brigade-15.js](../../examples/brigade-15.js)

In the script above, jobs `one` and `two` should each write a line to the file
`hello.txt`, which is stored in the shared `/mnt/brigade/share` directory. Since this
directory is shared among all three jobs, when the third job runs, it should print out the
results of the other two jobs.

So we can check the output of the third job's log:

```console
$ kubectl logs three-1504091079871-master
hello
world
```

That is exactly what we would expect to see.

Importantly, shared storage space is limited to 50 megabytes of storage per build. This can
be overridden in the project configuration.

> Note: Shared storage is dependent on the underlying Kubernetes cluster. Some Kubernetes
> clusters cannot support dynamically provisioned PVCs. If you run into problems with
> this, consult your Kubernetes admin or the Kubernetes storage documentation.

### Job Caches

The shared storage we saw above only persists as long as the build is running. When the
build is complete, the storage is recycled.

Brigade provides a second kind of storage that is designed to improve the speed of individual
jobs by giving them access to a cache.

A _job cache_ provides a place for a job to store data that it can access every time it is
run. This can drastically improve performance for things like dependency caching (during
code builds) or indexing.

A job cache is not enabled by default. A job must declare that it needs a cache.

```javascript
const { events, Job } = require("brigadier")

events.on("exec", (e) => {
  var job = new Job("cacher", "alpine:3.4")
  job.cache.enabled = true

  job.tasks = [
    "echo " + e.buildID + " >> /mnt/brigade/cache/jobs.txt",
    "cat /mnt/brigade/cache/jobs.txt"
  ]

  job.run()
})
```
[brigade-16.js](../../examples/brigade-16.js)

This script creates a new job and then enables the cache. Then it runs two different
tasks. The first writes the build's unique ID into a file in the cache, and the second one
echos the contents of that generated file.

If we run the above a few times and then check the job's output, we'll see one ID for each
time the job was run.

```console
$ kubectl logs cacher-1504091963651-master
brigade-worker-01brzvq79mj4849vy7ae0fez44-master
brigade-worker-01brzvqnhz21yjjh8m0zyxrxsc-master
brigade-worker-01brzvrhwe50xtktf4gcqk94wj-master
```

This happens because each time the job runs, the new build ID is written into the same
file that the job used on other builds.

By default, job caches are limited to only 5 megabytes (`5Mi`). However, you can easily
change this by setting `job.cache.size` to a larger value (`50Mi`, `5Gi`).

There are two final observations to make about job caches:

1. While a job cache is designed to persist across multiple runs, they are still
   considered to be _volatile storage_, which means a cache can reset. Do not use it as if
   it were stable long-term storage.
2. Caches are not automatically destroyed by Brigade (though other systems may clean them).
   That means that if you add lots and lots of jobs with caches enabled, lots of storage
   space will be reserved even if it is unused.

### Docker Runtime

Each job has the option to mount in a docker socket. When enabled, a docker socket is mounted to
`/var/run/docker.sock` in the job's container.

In order for the socket to be mounted, the Brigade project must have `Allow host mounts` set to
`true`. This can be set via the Advanced Configuration Options during `brig create`.  (See the [projects](projects.md)
doc for further info.)

This is a toggle-able option for all jobs, but is not enabled by default. A job must declare that it
needs a docker socket.

```javascript
const { events, Job } = require("brigadier")

events.on("exec", (e, p) => {
  var one = new Job("one", "docker:17.06.0-ce", ["docker images"])

  one.docker.enabled = true

  one.run()
})
```

This script creates a new job and then enables docker. It runs a `docker` command to demonstrate
what containers are available in this engine.

> Note: The Docker Runtime mounts in the host's Docker socket. Kubernetes administrators may not
> want to give users direct access to the kubelet's Docker daemon. If you're one of these people,
> you can disable jobs from being able to mount this by disabling it in the project's settings.
> If you're using the default project settings, this feature is disabled.

#### Docker-in-Docker

For security reasons, it is recommended that you use Docker-in-Docker (DinD) instead of using
the Docker socket directly.

DinD containers must run as privileged in order to function.  Therefore, the Brigade project must have 
`Allow privileged jobs` set to `true`.  This can be set via the Advanced Configuration Options
during `brig create`.  (See the [projects](projects.md) doc for further info.)

Here is an example job definition that uses the official Docker image to do a Docker
build:

```javascript
var driver = project.secrets.DOCKER_DRIVER || "overlay"

// Build and push a Docker image.
const docker = new Job("dind", "docker:stable-dind")
docker.privileged = true;
docker.env = {
  DOCKER_DRIVER: driver
}
docker.tasks = [
  "dockerd-entrypoint.sh &",
  "sleep 20",
  "cd /src",
  "docker pull brigadecore/kashti:canary || true",
  "docker build -t brigadecore/kashti:canary ."
];

docker.run()
```

This method is slower than using the Docker socket directly, but it is safer.

The Kashti `brigade.js` does a Docker build, and then (if configured) a push to the
upstream Docker registry:

```javascript
var driver = project.secrets.DOCKER_DRIVER || "overlay"

// Build and push a Docker image.
const docker = new Job("dind", "docker:stable-dind")
docker.privileged = true;
docker.env = {
  DOCKER_DRIVER: driver
}
docker.tasks = [
  "dockerd-entrypoint.sh &",
  "sleep 20",
  "cd /src",
  "docker pull brigadecore/kashti:canary || true",
  "docker build -t brigadecore/kashti:canary ."
];

// If a Docker user is specified, we push.
if (project.secrets.DOCKER_USER) {
  docker.env.DOCKER_USER = project.secrets.DOCKER_USER
  docker.env.DOCKER_PASS = project.secrets.DOCKER_PASS
  docker.env.DOCKER_REGISTRY = project.secrets.DOCKER_REGISTRY
  docker.tasks.push("docker login -u $DOCKER_USER -p $DOCKER_PASS $DOCKER_REGISTRY")
  docker.tasks.push("docker push brigadecore/kashti:canary")
} else {
  console.log("skipping push. DOCKER_USER is not set.");
}
docker.run()
```

In the above, if the docker credentials are set for the project, a `docker push`
is performed on the image just built.

#### Attaching volumes and volume mounts to jobs

Build storage and job cache represent a very simple and convenient way that Brigade exposes
for attaching storage to your jobs. But if your job requires the mounting of an existing 
[Kubernetes volume](https://kubernetes.io/docs/concepts/storage/volumes/), the Brigade JavaScript 
API exposes two propeties on the `Job` class:

- `volumes`: list of Kubernetes volumes to be attached to the pod specification. Supports all [Kubernetes 
volume types](https://kubernetes.io/docs/concepts/storage/volumes/#types-of-volumes) supported by your cluster configuration. Volumes are referenced by name in the `volumeMounts` property.
To reference a volume of type `hostPath`, the Brigade project must allow host mounts. 
- `volumeMounts`: list of Kubernetes volume mounts to be attached to all containers in the job pod specification, referenced 
by their names. Volumes referenced here must be defined in the `volumes` property.

> Note: simple use cases for build storage or job caches should still use the existing Brigade properties for 
> enabling the storage and cache. These properties should only be used in advanced scenarios that require mounting 
> Kubernetes volumes.

> This functionality was introduced with Brigade 1.2, and is in experimental state.

Example:

```javascript
    var j = new Job("some-image");
    j.volumes = [
        {
            name: "modules",
            hostPath: {
                path: "/lib/modules",
                type: "Directory"
            }
        },
        {
            name: "cgroup",
            hostPath: {
                path: "/sys/fs/cgroup",
                type: "Directory"
            }
        },
        {
            name: "docker-graph-storage",
            emptyDir: {}
        }
    ];

    j.volumeMounts = [
        {
            name: "modules",
            mountPath: "/lib/modules",
            readOnly: true
        },
        {
            name: "cgroup",
            mountPath: "/sys/fs/cgroup"
        },
        {
            name: "docker-graph-storage",
            mountPath: "/var/lib/docker"
        }
    ];
```

## Jobs and Return Values

We have seen already that when we run a job, it will return a JavaScript Promise.
Each Promise also wraps a value. And the value is the output of the job. So it is possible
to run a job, capture its output, and use that as input to the next job.

To illustrate this, let's write a script that creates two jobs. The first job will
simply echo a value. Then we will capture that value and send it to the second job.
Then we will capture the output of the second job and write it directly out to
to the console.

```javascript
const { events, Job } = require("brigadier")

events.on("exec", (e, p) => {
  var one = new Job("one", "alpine:3.4")
  var two = new Job("two", "alpine:3.4")

  one.tasks = ["echo world"]
  one.run().then( result => {
    two.tasks = ["echo hello " + result.toString()]
    two.run().then( result2 => {
      console.log(result2.toString())
    })
  })
})
```
[brigade-17.js](../../examples/brigade-17.js)

Our actual containers (`one` and `two`) are not doing much work here. They're just
echoing content to their standard output. But that information is captured by Brigade
and sent back as a `Result` object.

- Job `one` returns `world`.
- Job `two` takes that, and appends it to `echo hello`, which returns `hello world`
- So `result2` will have a `Result` with `hello world`

```
Started brigade-worker-01bsy7k5h6n65gtt5sfrstte99-master
yarn start v0.27.5
$ node prestart.js
prestart: src/brigade.js written
$ node --no-deprecation ./dist/src/index.js
Creating PVC named brigade-worker-01bsy7k5h6n65gtt5sfrstte99-master
looking up default/github-com-brigadecore-empty-testbed-one
Creating secret one-1505326897996-master
Creating Job Cache PVC github-com-brigadecore-empty-testbed-one
Creating pod one-1505326897996-master
Timeout set at 900000
default/one-1505326897996-master phase Pending
default/one-1505326897996-master phase Pending
default/one-1505326897996-master phase Succeeded
looking up default/github-com-brigadecore-empty-testbed-two
Creating secret two-1505326904108-master
Creating pod two-1505326904108-master
Creating Job Cache PVC github-com-brigadecore-empty-testbed-two
Timeout set at 900000
default/two-1505326904108-master phase Pending
default/two-1505326904108-master phase Succeeded
hello world

beforeExit 0
after handler is cleaning up build storage for brigade-worker-01bsy7k5h6n65gtt5sfrstte99-master
exit called
Done in 11.20s.
```

It is a good idea to call `toString()` on the `Result` object, since there is otherwise
no guarantee about the type of data returned.

With carefully constructed containers, you can get sophisticated and send structured
data like JSON from one job to another. But remember that what is captured is the
standard output (STDOUT) of the job, which is often where log data will also be sent.
Sometimes it makes more sense to write structured files to the shared storage instead.

## Advanced Event Handling

We have looked at ways of declaring simple event handlers, but it is possible to
chain together events. One event handler can trigger another event:

```javascript
const {events} = require("brigadier")

events.on("exec", function(e, project) {
  events.emit("next", e, project)
})

events.on("next", () => {
  console.log("fired 'next' event")
})
```
[brigade-18.js](../../examples/brigade-18.js)

The example above uses `events.emit` to fire a new event. In that example, we
re-use an existing event, which is not necessarily the best practice. A more
reliable way of triggering an event is to create a new `BrigadeEvent` and fire
that event:

```javascript
const {events} = require("brigadier")

events.on("exec", function(e, project) {
  const e2 = {
    type: "next",
    provider: "exec-handler",
    buildID: e.buildID,
    workerID: e.workerID,
    cause: {event: e}
  }
  events.fire(e2, project)
})

events.on("next", (e) => {
  console.log(`fired ${e.type} caused by ${e.cause.event.type}`)
})
```
[brigade-19.js](../../examples/brigade-19.js)

In this example, `e2` is a new event. Any new event _must_ have the following fields:

- `type`: The name of the event to fire
- `provider`: A name to indicate what fired the event
- `buildID`: The Build ID
- `workerID`: The Worker ID

When chaining events, it is considered good practice to also include the original
event as `cause: {event: e}`.

It is also possible to register more than one event handler for a single event.
In such cases, _all_ of the matching event handlers will be called.

```javascript
const {events} = require("brigadier")

events.on("exec", () => {
  console.log("first")
})

events.on("exec", () => {
  console.log("second")
})
```
[brigade-20.js](../../examples/brigade-20.js)

In this case, when the `exec` event is run, _both handlers will execute in the
order they are defined_.

Finally, it is also possible to nest event handlers so that one handler is only
registered when another event is called.

```javascript
const {events} = require("brigadier")

events.on("exec", (e, project) => {
  // This is only registered when 'exec' is called.
  events.on("next", () => {
    console.log("fired 'next' event")
  })
  events.emit("next", e, project)
})

events.on("exec2", (e, project) => {
  events.emit("next", e, project)
})
```
[brigade-21.js](../../examples/brigade-21.js)

In the example above, the `next` event handler is _only registered_ if the
`exec` event is run. Triggering the event `exec` will also trigger the wrapped
`next` handler. But triggering `exec2` will NOT trigger the defined `next` handler.

> Note that `events.on("next"...)` must be specified before `events.emit("next"...)`
> is called.

This particular feature can be useful when writing `after` and `error` handlers
(neither of which need direct invocations with `emit` or `fire`).

## Conclusion

This guide covers the basics of working with `brigade.js` files. If you are not sure how to
get started with Brigade now, you might want to take a look at the [tutorial](../intro). If
you want more details about the Brigade JS API, you can look at the [API
documentation](javascript.md). For a more advanced scripting guide, check [here](scripting_advanced.md).

Happy Scripting!
