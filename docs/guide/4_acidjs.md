# Creating an Acid.js File

Now that we have configured an Acid project, it's time to make use of it.

An `acid.js` file must be placed in the root of your Git repo and committed.

Acid uses simple JavaScript files to run tasks. When it comes to task running,
Acid follows this process:

- Listen for an event
- When the event is fired, execute the event handler (if found) in `acid.js`
- Wait until the event is handled, then report the result

Given this, the role of the `acid.js` file is to declare event handlers. And it's
easy.

```javascript
events.push = function(e) {
  console.log("received push for commit " + e.commit)
}
```

The above defines one event: `push`. This event responds to GitHub `push` requests
(like `git push origin master`). If you have configured your GitHub webhook system
correctly (see previous section) then each time GitHub receives a push, it will
notify Acid.

Acid will run the `events.push` event handler, and it will give that event handler
a single parameter (`e`), which is a record of the event that was just triggered.

In our script above, we just log the comment:

```
received push for commit e459558...
```

Note that `e.commit` holds the Git commit SHA for the commit that was just pushed.

# Adding a Job

Logging a commit SHA isn't all that helpful. Instead., let's imagine that our
project is a Node.js project, and we want to build and test it. Now we want to
add a build job:

```javascript
events.push = function(e) {
  // A slightly better log message.
  console.log("===> Building " + e.repo.cloneURL + " " + e.commit);

  // Create a new job
  var node = new Job("node-runner")

  // We want our job to run the stock Docker node:8 image (NodeJS version 8)
  node.image = "node:8"

  // Now we want it to run these commands in order:
  node.tasks = [
    "cd /src/hello",
    "npm install",
    "npm run test"
  ]

  // We're done configuring, so we run the job
  node.run()
}
```

The example above is derived from https://github.com/deis/empty-testbed

The example above introduces Acid `Job`s. A Job is a particular build step. Each
job can run a Docker container and feed it multiple commands.

Above, we create the `node-runner` job, have it use the [node:8](https://hub.docker.com/_/node/)
image, and then set it up to run the following commands in that container:

- `cd /src/hello`: CD into the directory that contains the source code. This is a Git checkout of your code.
- `npm install`: Use the NPM.js installer to configure stuff for us.
- `npm run test`: Run the unit tests for our project.

Finally, when we run `node.run()`, the job is built and executed. If it passes,
all is good. If it fails, Acid is notified (and if you set up GitHub notifications,
GitHub is notified as well).

# Moving Forward

From here, you have the tools to begin writing Acid builds. You can read the
[Acid JavaScript](../javascript) to learn about the rest of the API, or you can
look at the configuration guide **NOT CREATED** to learn about more configuration
options.

But here are a few things to help you:

- You can run as many jobs as you like in each event handler. We did just one, but
  you can do many.
- Feel free to use functions and objects in your scripts if you want to break
  things down into smaller parts.
- To get advanced flow control, take a look at `waitgroup`, which allows you to
  run a batch of jobs concurrently, then wait for them all to finish.
- DO NOT try to load external JavaScript modules (like Node modules). This is not
  supported.
