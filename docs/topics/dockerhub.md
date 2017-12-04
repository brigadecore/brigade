# Container Registry (DockerHub, ACR) Integration

Brigade supports container registry webhooks such as the ones emitted by
DockerHub and ACR. The following platforms are known to work with
Brigade's Container Registry gateway:

- DockerHub
- Azure Container Registry (ACR) with the `Managed_*` classes

DockerHub/ACS integration is _not enabled by default_.

## Intro to Container Registry Webhook Integration

Brigade comes with built-in support for container registry image pushing events. When a
container registry webhook system is configured to notify Brigade's GW server, Brigade will
respond to an image push by triggering an `image_push` event.

This provides Brigade developers with the ability to trigger scripts based on a
new image being pushed to a Docker repository.

## Configuring Brigade

Container Registry support is disabled by default, but can easily be turned on
during installation or upgrade of Brigade:

```
$ helm install -n brigade brigade/brigade --set cr.enabled=true
```

This will enable the container registry. You will likely also want to expose the
container registry outside of the cluster so that inbound webhooks will work. The
easiest way to do this is to also set up a service of type `LoadBalancer`:

```
$ helm install -n brigade brigade/brigade --set cr.enabled=true,cr.service.type=LoadBalancer
```

A more secure route is to install an SSL proxy (like `kube-lego`) and directing
that to the internal container registry service.

For more installation configuration options, run `helm inspect values brigade/brigade`
and read the `cr:` section.

## Configuring the Repository

The repository _must_ support web hooks.

The URL pattern for calling a webhook is this:

```
http://<YOUR GATEWAY>:7744/events/dockerhub/<YOUR  PROJECT NAME>/<COMMIT>
```

For example, to connect to the project `technosophos/example-hook` and use the head
commit, we would use:

```
http://technosophos.brigade.sh:7744/events/dockerhub/technosophos/example-hook/master
```

For DockerHub, this URL is added in the `webhooks` tab of the Docker repository for
your image.

For Azure Container Registry, this URL is added on the `webhooks` tab of your
ACR repository's blade.

## Configuring your `brigade.js`

To answer hooks in your `brigade.sh`, you will need to do something like this:

```javascript
const {events, Job} = require("brigadier")

events.on("image_push", (e, p) => {
  var docker = JSON.parse(e.payload)
  console.log(docker)
})
```

The webhook data sent by DockerHub is different than the data sent by Azure
Container Registry. The following example uses ACR's `action` and `target`
objects:

```javascript
const {events, Job} = require("brigadier")

events.on("image_push", (e, p) => {
  var docker = JSON.parse(e.payload)

  // Currently the only action sent is 'push', but this makes your script
  // safe for the future.
  if (docker.action != "push") {
    console.log(`ignoring action ${docker.action}`)
    return
  }

  // Here's how you get the tag.
  var version = docker.target.tag || "latest"
  console.log(`image version: ${version}`)
}
```

The above answers an ACR webhook. The data sent by DockerHub's webhook is
[slightly different](https://docs.docker.com/docker-hub/webhooks/).

**IMPORTANT:** An event will trigger for _every tag you push_, even if that tag
is not new or updated. If you push both a `latest` and a versioned tag for a
single image, you will get two webhook invocations.


