# Generic Gateway

Brigade contains a Generic Gateway that can be used to accept requests from other platforms or systems. Generic Gateway is a separate component in the Brigade system, like Github and Container Registry (CR) Gateways.

Generic Gateway is _not enabled by default_.

## Intro to Generic Gateway 

Generic Gateway can optionally be activated to accept `POST` webhook requests at `/webhook/:projectID/:secret` path. When this endpoint is called, Brigade will respond by creating a Build with a `webhook` event. This provides Brigade developers with the ability to trigger scripts based on messages received from any platform that can send a POST HTTP request.

## Generic Gateway

Generic Gateway is disabled by default, but can easily be turned on during installation or upgrade of Brigade:

```
$ helm install -n brigade brigade/brigade --set genericGateway.enabled=true
```

This will enable the Generic Gateway Deployment/Services/RBAC permissions. However, be aware that Generic Gateway is not exposed outside the cluster. In case you are certain from a security perspective that you want to expose the Generic Gateway via a [Kubernetes LoadBalancer Service](https://kubernetes.io/docs/concepts/services-networking/#loadbalancer), you can install Brigade via the following command:

```
$ helm install -n brigade brigade/brigade --set genericGateway.enabled=true,genericGateway.service.type=LoadBalancer
```

Alternatively, for enhanced security, you can install an SSL proxy (like `cert-manager`) and direct it to the Generic Gateway Service.

## Using the Generic Gateway

As mentioned, Generic Gateway accepts POST requests at `/webhook/:projectID/:secret` endpoint. These requests should also carry a JSON payload.

- `projectID` is the Brigade Project ID
- `secret` is a custom secret for this specific project's Generic Gateway webhook support. In other words, each project that wants to accept Generic Gateway events should have its own Generic Gateway secret. This secret serves as a simple authentication mechanism.

When you create a new Brigade Project via Brig CLI, you can optionally create such a secret by using the Advanced Options during `brig project create`. This secret must contain only alphanumeric characters. If you provide an empty string, Brig will generate and output a secret for you.

*Important*: If you do not go into "Advanced Options" during `brig project create`, a secret will not be created and you will not be able to use Generic Gateway for your project. However, you can always use `brig project create --replace` (or just `kubectl edit` your project Secret) to update your project and include a `genericGatewaySecret` string value.

When calling the Generic Gateway webhook endpoint, you can include a custom JSON payload such as:

```json
{
	"ref": "refs/heads/changes",
	"commit": "b60ad9543b2ddbbe73430dd6898b75883306cecc"
}
```

`Ref` and `commit` values would be used to configure the specific revision that Brigade will pull from your repository.

Last but not least, here is a sample Brigade.js file that could be used as a base for your own scripts that respond to Generic Gateway's `webhook` event. This script will echo the name of your project and 'webhook'.

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

---

Prev: [Container Registry Integration](dockerhub.md) `|` Next: [Using Secrets](secrets.md)

Return to the [Table of Contents](index.md)