# Generic Webhook

Brigade contains a Generic Webhook which is part of its API Server and can be used to accept requests from other platforms or systems.

Generic Webhook is _not enabled by default_.

## Intro to Generic Webhook 

Brigade API Server can optionally be activated to accept `POST` webhook requests at `/webhook/:projectID/:secret` path. When this endpoint is called, Brigade will respond by creating a Build with a `webhook` event. This provides Brigade developers with the ability to trigger scripts based on messages received from any platform that can send a POST HTTP request.

## Configuring Brigade API Server for Generic Webhook

Generic Webhook support is disabled by default, but can easily be turned on during installation or upgrade of Brigade:

```
$ helm install -n brigade brigade/brigade --set genericGateway.enabled=true
```

This will enable the Generic Webhook on Brigade API Server. However, be aware that Brigade API Server is not exposed outside the cluster. In case you are certain from a security perspective that you want to expose the entire API Server via a [Kubernetes LoadBalancer Service](https://kubernetes.io/docs/concepts/services-networking/#loadbalancer), you can install Brigade via the following command:

```
$ helm install -n brigade brigade/brigade --set genericGateway.enabled=true,api.service.type=LoadBalancer
```

Alternatively, for enhanced security, you can install an SSL proxy (like `cert-manager`) and direct it to the internal API Server service.

## Using the Generic Webhook

As mentioned, Generic Webhook accepts POST requests at `/webhook/:projectID/:secret` endpoint. These requests should also carry a JSON payload.

- `projectID` is the Brigade Project ID
- `secret` is a custom secret for this specific project's Generic Webhook support. You can think of it as a simple authentication mechanism for this Project's Generic Webhook support. Every Project has its own unique secret.

When you create a new Brigade Project via Brig CLI, you can optionally create such a secret by adding a secret named `genericWebhookSecret`, containing your desired value. Alternatively, Brig will generate and output one for you.

When calling the Generic Webhook endpoint, you can include a custom JSON payload such as this one:

```json
{
	"ref": "refs/heads/changes",
	"commit": "b60ad9543b2ddbbe73430dd6898b75883306cecc"
}
```

`Ref` and `commit` values would be used to configure the specific revision that Brigade will pull from your repository.

Last but not least, here is a sample Brigade.js file that could be used as a base for your own scripts that respond to Generic Webhook's `webhook` event. This script will echo the name of your project and 'webhook'.

```javascript
const { events, Job } = require("brigadier");
events.on("webhook", (e, p) => {
  var echo = new Job("echo", "alpine:3.8");
  echo.storage.enabled = false;
  echo.tasks = [
    "echo Project " + p.name,
    "echo Event $EVENT_NAME"
  ];

  echo.env = {
    "EVENT_NAME": e.type
  };

  echo.run();
});
```