---
title: 'Workers'
description: 'How to add custom libraries to a Brigade worker'
---

# Adding custom libraries to a Brigade worker

A worker is the component in Brigade that executes a `brigade.js`. Brigade ships
with a general purpose worker focused on running jobs. This generic worker
exposes a host of Node.js libraries to `brigade.js` files, as well as the
`brigadier` Brigade library.

Sometimes, though, it is desirable to include additional libraries -- perhaps even
custom libraries -- to your workers. 
There are two ways of adding custom dependencies to a Brigade worker:

- by creating a custom Docker image for the worker that already contains the dependencies 
and which will be used for all Brigade projects.

- without creating a custom container image - check [the dependencies document](dependencies.md) for a detailed description for this approach.

**Note:** This area of Brigade is still evolving. If you have ideas for improving
it, feel free to [file an issue](https://github.com/Azure/brigade/issues) explaining
your idea.


## Workers and Docker Images

The Brigade worker (`brigade-worker`) is captured in a Docker image, and that
image is then executed in a container on your cluster. Augmenting the worker,
then, is done by creating a custom Docker image and then configuring Brigade
to use that image instead of the default `brigade-worker` image.

Next we will see how to quickly create a custom worker by creating a new
Docker image based on the base image.

## Creating a Custom Worker

As we saw above, workers are Docker images. And the default worker is a Docker
image that contains the Brigade worker runtime, which can read and execute
`brigade.js` files.

At its core, the Brigade worker is just a Node.js application. That means that
we can use the usual array of Node.js tools and libraries. Here, we'll load an
extra library from NPM so that it is available inside of our `brigade.js`
scripts.

Since the main worker is already tooled to do the main processing, the easiest
way to add your own libraries is to start with the existing worker and add to
it. Docker makes this convenient.

Say we want to provide an XML parser to our Brigade scripts. We can do that
using a `Dockerfile`:

```Dockerfile
FROM deis/brigade-worker:latest

RUN yarn add xml-simple
```

The above will begin with the default Brigade worker and simply add the `xml-simple`
library. We can build this into an image and then push it to our Docker registry
like this:

```console
$ docker build -t myregistry/myworker:latest .
$ docker push myregistry/myworker:latest
```

> IMPORTANT: Make sure you replace `myregistry` and `myworker` with your own
> account and image names.

**Tip:** If you are running a local Kubernetes cluster with Docker or Minikube,
you do not need to push the image. Just configure your Docker client
to point to the same Docker daemon that your Kubernetes cluster is using. (With
Minikube, you do this by running `eval $(minikube docker-env)`.)

Now that we have our image pushed to a usable location, we can configure Brigade
to use this new image.

## Configuring Brigade to Use Your New Worker

As of Brigade v0.10.0, worker images can be configured _globally_. Individual
projects can choose to override the global setting.

To set the version globally, you should override the following values in your
`brigade/brigade` chart:

```yaml
# worker is the JavaScript worker. These are created on demand by the controller.
worker:
  registry: myregistry
  name: myworker
  tag: latest
  #pullPolicy: IfNotPresent # Set this to Always if you are testing and using
  #                           an upstream registry like Dockerhub or ACR
```

You can then use `helm upgrade` to load those new values to Brigade.

### Project Overrides

To configure the worker image per-project, you can set up a custom `worker` section
via `brig` during the `Configure advanced options` section.  (If the project has
already been created, use `brig project create --replace -p <pre-existing-project>`).

Here we supply our custom worker image registry (`myregistry`), image name
(`myworker`), image tag (`latest`), pull policy (`IfNotPresent`) and command (`yarn -s start`):

```console
$ brig project create
...
? Configure advanced options Yes
...
? Worker image registry or DockerHub org myregistry
? Worker image name myworker
? Custom worker image tag latest
? Worker image pull policy IfNotPresent
? Worker command yarn -s start
```

## Using Your New Image

Once you have set the Docker image (above), your new Brigade workers will
automatically switch to using this new image.

Assuming you have configured your project (as explained above) and
your Kubernetes cluster can see the Docker registry that you pushed your image to,
you can now simply assume that you are using your new custom image. So now
we can import our new `simple-xml` library:

[brigade.js](examples/workers/brigade.js)
```javascript
const { events } = require("brigadier");
const XML = require("simple-xml");

events.on("exec", () => {
  var o = XML.parse("<say><to>world</to></say>")
  console.log(`Saying hello to ${o.say.to}`);
});

```

Running the above with `brig run -f brigade.js my/project` (where `my/project`
is some project you have already created) should result in a successful run.

Here is an example:
```console
$ brig run -f brigade.js deis/empty-testbed
Started build 01c7kmserwyc5y05rrhpvnp5m0 as "brigade-worker-01c7kmserwyc5y05rrhpvnp5m0-master"
prestart: src/brigade.js written
[brigade] brigade-worker version: 0.10.0
[brigade:k8s] Creating PVC named brigade-worker-01c7kmserwyc5y05rrhpvnp5m0-master
Saying hello to world
[brigade:app] after: default event fired
[brigade:app] beforeExit(2): destroying storage
[brigade:k8s] Destroying PVC named brigade-worker-01c7kmserwyc5y05rrhpvnp5m0-master
```

You can see that `Saying hello to ${o.say.to}` rendered correctly
as `Saying hello to world`.

## Adding a Custom (Non-NPM) Library

Sometimes it is useful to encapsulate commonly used Brigade code into a library
that can be shared between projects internally. While the NPM model above is
easier to manage over the longer term, there is a simple method for loading
custom code into an image. This section illustrates that method.

Here is a small library that adds an `alpineJob()` helper function:

[mylib.js](examples/workers/mylib.js)
```javascript
const {Job} = require("./brigadier");

exports.alpineJob = function(name) {
  j = new Job(name, "alpine:3.7", ["echo hello"])
  return j
}
```

**Note:** Because we are loading our code directly into Brigade, we import
`./brigadier`, not `brigadier`. This may change in the future.

We can build this file into our Dockerfile by copying it into the image:

```
FROM deis/brigade-worker:latest

RUN yarn add xml-simple
COPY mylib.js /home/src/dist
```

And now we can build the above:

```console
$ docker build -t myregistry/myworker:latest .
$ docker push myregistry/myworker:latest
```

Assuming you have configured Brigade to use your `myworker` image (explained
above in "Configuring Brigade to Use Your New Worker"), you can begin using the
library:

```javascript
const { events } = require("brigadier");
const XML = require("xml-simple");
const { alpineJob } = require("./mylib");

events.on("exec", () => {
  XML.parse("<say><to>world</to></say>", (e, say) => {
    console.log(`Saying hello to ${say.to}`);
  })

  const alpine = alpineJob("myjob");
  alpine.run();
});

```

Now we've added a few new lines to the script. We import `alpineJob` from our
`./mylib` module at the top. Then, late in the script, we call `alpineJob` to
create a new job for us.

## Best practices

We strongly discourage attempting to turn a worker into a long-running server.
This violates the design assumptions of Brigade, and can result in unintended
side effects.
