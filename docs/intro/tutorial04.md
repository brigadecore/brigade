# Writing your first CI pipeline, Part 4

This tutorial begins where [Tutorial 3][part3] left off. We’ll walk through the process for writing your first feature for our UUID generator app, then test the feature on Github using Acid.

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

class AppTestCase(unittest.TestCase):

    def setUp(self):
        self.client = app.app.test_client()

    def test_uuid_generated(self):
        resp = self.client.get('/')
        assert uuid.UUID(resp.data)
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

## Create an acid.js file

Now that we have successfully written tests for our app and configured an Acid project, it's time to make use of them.

An `acid.js` file must be placed in the root of your git repo and committed.

Acid uses simple JavaScript files to run tasks. When it comes to task running, Acid follows this process:

- Listen for an event
- When the event is fired, execute the event handler (if found) in `acid.js`
- Wait until the event is handled, then report the result

Given this, the role of the `acid.js` file is to declare event handlers. And it's easy. Open `acid.js` and write this JavaScript code into it:

```javascript
events.on("push", function(e, project) {
  console.log("received push for commit " + e.commit)
})
```

The above defines one event: `push`. This event responds to Github `push` requests (like `git push origin master`). If you have configured your Github webhook system correctly (see [part 3][part3]) then each time GitHub receives a push, it will notify Acid.

Acid will run the `events.push` event handler, and it will give that event handler a single parameter (`e`), which is a record of the event that was just triggered.

In our script above, we just log the comment:

```
received push for commit e459558...
```

Note that `e.commit` holds the git commit SHA for the commit that was just pushed.

# Add a job

Logging a commit SHA isn't all that helpful. Instead, we would want to test that our UUID generator project is actually generating UUIDs, wouldn't we?

Edit `acid.js` again so it looks like this:

```javascript
events.on("push", function(e, project) {
  console.log("received push for commit " + e.commit)

  // Create a new job
  var node = new Job("test-runner")

  // We want our job to run the stock Docker Python 3 image
  node.image = "python:3"

  // Now we want it to run these commands in order:
  node.tasks = [
    "cd /src/",
    "pip install -r requirements.txt",
    "python setup.py test"
  ]

  // We're done configuring, so we run the job
  node.run()
})
```

The example above introduces Acid `Job`s. A Job is a particular build step. Each job can run a Docker container and feed it multiple commands.

Above, we create the `test-runner` job, have it use the [python:3](https://hub.docker.com/_/python/) image, and then set it up to run the following commands in that container:

- `cd /src/`: change into the directory that contains the source code
- `pip install -r requirements.txt`: Use pip to install Flask like we did in [part 1][part1].
- `python setup.py test`: Run the test suite for our project.

Finally, when we run `node.run()`, the job is built and executed. If it passes, all is good. If it fails, Acid and Github are notified.

This concludes the basic tutorial. If you are familiar with Acid and are interested in learning how to refactor acid.js into a more efficient test pipeline, check out [Advanced tutorial: Writing efficient pipelines][efficient-pipelines].

You might also be scratching your head on what to [read next][readnext].


[efficient-pipelines]: writing-efficient-pipelines.md
[part1]: tutorial01.md
[part3]: tutorial03.md
[readnext]: readnext.md
