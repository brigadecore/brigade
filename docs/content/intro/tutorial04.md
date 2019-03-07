---
title: 'Tutorial 4: Writing a Test'
description: 'Writing your first CI pipeline, Part 4'
section: intro
---

# Writing your first CI pipeline, Part 4

This tutorial begins where [Tutorial 3][part3] left off. We’ll walk through the process for writing your first feature for our UUID generator app, then test the feature on Github using Brigade.

## Test the application

Let’s check that the application shows a UUID if we access the root of the application (/). Let's create a test directory to write our tests:

```
$ mkdir tests/
$ touch tests/__init__.py
```

In order to test the application, open `tests/app_tests.py` and create a unittest skeleton there:

```python
import unittest
import uuid

import app

# Helper

def bytes_to_str(b):
    return ''.join(chr(x) for x in (b))

class AppTestCase(unittest.TestCase):

    def setUp(self):
        self.client = app.app.test_client()


    def test_uuid_generated(self):
        resp = self.client.get('/')
        assert uuid.UUID(bytes_to_str(resp.data))
```

Now we can run the test using a minimal basic setup script using setuptools, a built-in Python package that allows developers to more easily build and distribute Python packages.

Open `setup.py` and write this python code in there:

```python
from setuptools import setup, find_packages

setup(
    name="uuid-generator",
    version="0.1",
    packages=find_packages(),
    test_suite="tests",
)
```

This tells python to run the tests found in the tests/ directory we created earlier.

To run the tests, invoke

```
$ python setup.py test
running test
running egg_info
writing uuid_generator.egg-info/PKG-INFO
writing top-level names to uuid_generator.egg-info/top_level.txt
writing dependency_links to uuid_generator.egg-info/dependency_links.txt
reading manifest file 'uuid_generator.egg-info/SOURCES.txt'
writing manifest file 'uuid_generator.egg-info/SOURCES.txt'
running build_ext
test_uuid_generated (tests.app_tests.AppTestCase) ... ok

----------------------------------------------------------------------
Ran 1 test in 0.010s

OK
```

Now is a good time to commit your work.

```
$ git add tests/ setup.py
$ git commit -m "add unit tests"
$ git push origin master
```

## Create a brigade.js file

Now that we have successfully written tests for our app and configured a Brigade project, it's time to make use of them.

An `brigade.js` file must be placed in the root of your git repo and committed.

Brigade uses simple JavaScript files to run tasks. When it comes to task running, Brigade follows this process:

- Listen for an event
- When the event is fired, execute the event handler (if found) in `brigade.js`
- Wait until the event is handled, then report the result

Given this, the role of the `brigade.js` file is to declare event handlers. And it's easy. Open `brigade.js` and write this JavaScript code into it:

```javascript
const { events } = require("brigadier");

events.on("push", function(e, project) {
  console.log("received push for commit " + e.revision.commit)
})
```

The above defines one event: `push`. This event responds to Github `push` requests (like `git push origin master`). If you have configured your Github webhook system correctly (see [part 3][part3]) then each time GitHub receives a push, it will notify Brigade.

Brigade will run the `events.push` event handler, and it will give that event handler a single parameter (`e`), which is a record of the event that was just triggered.

In our script above, we just log the comment:

```
received push for commit e459558...
```

Note that `e.revision.commit` holds the git commit SHA for the commit that was just pushed.

# Add a job

Logging a commit SHA isn't all that helpful. Instead, we would want to test that our UUID generator project is actually generating UUIDs, wouldn't we?

Edit `brigade.js` again so it looks like this:

```javascript
const { events, Job } = require("brigadier");

events.on("push", function(e, project) {
  console.log("received push for commit " + e.revision.commit)

  // Create a new job
  var testRunner = new Job("test-runner")

  // We want our job to run the stock Docker Python 3 image
  testRunner.image = "python:3"

  // Now we want it to run these commands in order:
  testRunner.tasks = [
    "cd /src/app",
    "pip install -r requirements.txt",
    "cd /src/",
    "python setup.py test"
  ]

  // Display logs from the job Pod
  testRunner.streamLogs = true;

  // We're done configuring, so we run the job
  testRunner.run()
})
```

The example above introduces Brigade jobs. A Job is a particular build step. Each job can run a Docker container and feed it multiple commands.

Above, we create the `test-runner` job, have it use the [python:3](https://hub.docker.com/_/python/) image, and then set it up to run the following commands in that container:

- `cd /src/`: change into the directory that contains the source code
- `pip install -r requirements.txt`: Use pip to install Flask like we did in [part 1][part1].
- `python setup.py test`: Run the test suite for our project.

Finally, when we run `testRunner.run()`, the job is built and executed. If it passes, all is good. If it fails, Brigade and Github are notified.

## Wrapping our Job in GitHub Checks

To notify GitHub of our test results utilizing the [Checks API](https://developer.github.com/v3/checks/),
we'll need to add a bit more to our `brigade.js` file. 

```javascript
const { events, Job } = require("brigadier");

// GitHub Check events to watch for
//
// Note that a GitHub App will automatically generate these events
// from a `push` event, so we don't need an explicit push event handler any longer
events.on("check_suite:requested", checkRequested);
events.on("check_suite:rerequested", checkRequested);
events.on("check_run:rerequested", checkRequested);

// Our main test logic, refactored into a function that returns the job
function runTests(e, project) {
  // Create a new job
  var testRunner = new Job("test-runner");

  // We want our job to run the stock Docker Python 3 image
  testRunner.image = "python:3";

  // Now we want it to run these commands in order:
  testRunner.tasks = [
    "cd /src",
    "pip install -r requirements.txt",
    "python setup.py test"
  ];

  // Display logs from the job Pod
  testRunner.streamLogs = true;

  return testRunner;
}

// This runs our main test job, updating GitHub along the way
function checkRequested(e, p) {
  console.log("check requested");

  // This Check Run image handles updating GitHub
  const checkRunImage = "deis/brigade-github-check-run:latest";

  // Common configuration
  const env = {
    CHECK_PAYLOAD: e.payload,
    CHECK_NAME: "Brigade",
    CHECK_TITLE: "Run Tests",
  };

  // For convenience, we'll create three jobs: one for each GitHub Check
  // stage.
  const start = new Job("start-run", checkRunImage);
  start.imageForcePull = true;
  start.env = env;
  start.env.CHECK_SUMMARY = "Beginning test run";

  const end = new Job("end-run", checkRunImage);
  end.imageForcePull = true;
  end.env = env;

  // Now we run the jobs in order:
  // - Notify GitHub of start
  // - Run the tests
  // - Notify GitHub of completion
  //
  // On error, we catch the error and notify GitHub of a failure.
  start.run().then(() => {
    return runTests(e, p).run()
  }).then( (result) => {
    end.env.CHECK_CONCLUSION = "success"
    end.env.CHECK_SUMMARY = "Build completed"
    end.env.CHECK_TEXT = result.toString()
    return end.run()
  }).catch( (err) => {
    // In this case, we mark the ending failed.
    end.env.CHECK_CONCLUSION = "failure"
    end.env.CHECK_SUMMARY = "Build failed"
    end.env.CHECK_TEXT = `Error: ${ err }`
    return end.run()
  });
}
```

At this point, you should commit your work to a new branch and check that it all works:

```
$ git checkout -b add-brigade
$ git add .
$ git commit -m "add brigade.js"
$ git push origin add-brigade
```

Open up a new pull request on your repository using the branch:

![New PR](https://docs.brigade.sh/img/img4.png)

And we should now see the Checks tab with updates/results from the resulting Check Suite:

![GitHub Checks](https://docs.brigade.sh/img/img5.png)

This concludes the basic tutorial. If you are familiar with Brigade and are interested in learning how to refactor brigade.js into a more efficient test pipeline, check out [Advanced tutorial: Writing efficient pipelines][efficient-pipelines].

You might also be scratching your head on what to [read next][readnext].


[efficient-pipelines]: ../writing-efficient-pipelines
[part1]: ../tutorial01
[part3]: ../tutorial03
[readnext]: ../readnext