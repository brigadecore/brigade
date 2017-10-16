# DockerHub Integration

Brigade supports DockerHub webhooks. The following platforms are known to work with
Brigade's DockerHub system:

- DockerHub
- Azure Container Registry (ACR) with the `Managed_*` classes

## Intro to DockerHub Integration

Brigade comes with built-in support for DockerHub image pushing events. When a
DockerHub webhook system is configured to notify Brigade's GW server, Brigade will
respond to an image push by triggering an `imagePush` event.

This provides Brigade developers with the ability to trigger scripts based on a
new image being pushed to a Docker repository.

## Configuring Brigade

Brigade comes with DockerHub integration out of the box. If the Brigade
gateway server is reachable by the image repository then you can use the system.

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

events.on("imagePush", (e, p) => {
  var docker = JSON.parse(e.payload)
  console.log(docker)
})
```

The webhook data sent by DockerHub is different than the data sent by Azure
Container Registry. The following example uses ACR's `action` and `target`
objects:

```javascript
const {events, Job} = require("brigadier")

events.on("imagePush", (e, p) => {
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


