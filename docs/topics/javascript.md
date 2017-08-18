# About the Acid JavaScript engine

Acid JavaScript is a dialect of JavaScript for writing Acid build files.

Acid JavaScript has access to a few libraries:

- The Underscore.js library is built into Acid.js
- AcidJS has a number of built-in objects.

## The `events` Global Object

There is a global variable called `events` that provides the event handlers for your Acid project. Attach your event handler to this object:

Acid will call events when they occur in the project's lifecycle.

Here is how you handle a push request from GitHub (which is usually the main think you want to do):

```
events.push = function(e) {
  var j = new Job("my-job")
  //...
  j.run()
}
```

Every event handler function gets an `Event` (`e`) object.

Defined events:

- `events.push`: A new commit was pushed to the main project


### The `project` Global Object

There is a global object named `project` that contains information about the project that the acid script is running within.

These fields are primarily derived from the `acid-project` installation on your Kubernetes cluster. See the `acid-project`'s `values.yaml` file for more.

Properties:

  - `id`: The unique ID of the project
  - `name`: The project name, typically `org/name`.
  - `secrets`: A set of name/values pairs that are stored in a Kubernetes Secret.
    This is typically where protected data like tokens or passphrases are stored.
  - `kubernetes`: The object describing this project's Kubernetes settings
    - `kubernetes.namespace`: The namespace for this project
    - `kubernetes.vcsSidecar`: The VCS sidecar image/tag
  - `repo`: Information on the upstream repository (if available).
    - `repo.name`: The name of the repo (`org/name`)
    - `repo.cloneURL`: The URL that the VCS software can use to clone the reposiotry.
    - `repo.sshKey`: The SSH key that can be used to clone the repository (if applicable).
      This is often empty.

Secrets (`project.secrets`) are passed from the project configuration into a Kubernetes Secret, then injected into Acid.

So `helm install acid-project --set secrets.foo=bar` will add `foo: bar` to
`project.secrets`.

### The Event object

The Event object describes an event.

Properties:

- `type`: The event type (e.g. `push`)
- `provider`: The entity that caused the event (`github`)
- `commit`: The commit ID that this script should operate on.
- `payload`: The object received from the event trigger. For GitHub requests, its
  the data we get from GitHub.


### The Job object

To create a new job:

```javascript
j = new Job(name)
```

Parameters:

- A job name (alpha-numeric characters plus dashes).

Properties:

- `name`: The name of the job
- `image`: A Docker image with optional tag.
- `tasks`: An array of commands to run for this job
- `shell`: The terminal emulator that job tasks will be executed under. By default,
  this is /bin/sh
- `env`: Key/value pairs that will be injected into the environment. The key is
  the variable name (`MY_VAR`), and the value is the string value (`foo`)

It is common to pass data from the `e.env` Event object into the Job object as is appropriate:

```javascript
events.push = function(e) {
  j = new Job("example")
  j.env = { "DB_PASSWORD": project.secrets.dbPassword }
  //...
  j.run()
}
```

The above will make `$DB_PASSWORD` available to the "example" job's runtime.

Methods:

- `run()`: Run this job and wait for it to exit.
- `background()`: Run this job in the background.
- `wait()`: Wait for a backgrounded job to complete.

### The WaitGroup object

A WaitGroup is a tool for running multiple jobs in parallel. Create a WaitGroup, add jobs, and then run them all in parallel:

```
j1 = new Job("one")
j2 = new Job("two")

// Configure jobs...


Start two jobs in parallel and wait for both to complete.
wg = new WaitGroup()
wg.add(j1)
wg.add(j2)
wg.run()
```

The above will report success if both jobs run to completion.

Methods:

- `add(j: Job)`: Add a job
- `run()`: Run all jobs in parallel, wait for them to complete
- `wait()`: If the jobs are already running, wait for them. Don't start jobs, though.
  This is used for cases where you run `j.background()`, then add them to the waitgroup.

## Acid JS and ECMAScript (and browser-based JS)

Acid JS is ECMAScript 5. It has a few differences, though.

- The Regular Expression library is Go's regular expression library
- It does not provide `setTimeout` or `setInterval`
- Browser objects, like `window`, are not present
