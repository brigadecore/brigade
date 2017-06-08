# Acid JavaScript

Acid JavaScript is a dialect of JavaScript for writing Acid build files.

Acid JavaScript has access to a few libraries:

- The Underscore.js library is built into Acid.js
- AcidJS has a number of built-in objects.

## The `events` Global Variable

There is a global variable called `events` that provides the event handlers for
your Acid project. Attach your event handler to this object:

Acid will call events when they occur in the project's lifecycle.

Here is how you handle a push request from GitHub (which is usually the main
think you want to do):

```
events.github.push = function(e) {
  var j = new Job("my-job")
  //...
  j.run()
}
```

Every event handler function gets an `Event` (`e`) object.

Defined events:

- `events.github.push`: A new commit was pushed to the main project

### The Event object

The Event object describes an event.

Properties:

- `request`: The object received from the event trigger. For GitHub requests, its
  the data we get from GitHub.
- `config`: A dictionary of configuration name/value pairs.
- `name`: The name of the event (e.g. `github.push`)


### The Job object

To create a new job:

```javascript
j = new Job(name)
```

Parameters:

- A job name (alpha-numeric characters plus dashes).

Properties:

- `image`: A Docker image with optional tag.
- `env`: Key/value pairs that will be injected into the environment. The key is
  the variable name (`MY_VAR`), and the value is the string value (`foo`)
- `secrets`: Key/value pairs where the key is the name of the environment variable
  and the value is the name of the item in the Secret. `{ "DB_PASS": "dbpassword" }`

Methods:

- `run()`: Run this job and wait for it to exit.
- `background()`: Run this job in the background.
- `wait()`: Wait for a backgrounded job to complete.

### The WaitGroup object

A WaitGroup is a tool for running multiple jobs in parallel. Create a WaitGroup,
add jobs, and then run them all in parallel:

```
j1 = new Job("one")
j2 = new Job("two"

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
