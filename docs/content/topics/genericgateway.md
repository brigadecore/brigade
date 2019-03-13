---
title: Generic Gateway
description: About Brigade's generic gateway.
---

# Generic Gateway

Brigade contains a Generic Gateway that can be used to accept requests from other platforms or systems. Generic Gateway is a separate component in the Brigade system, like Github and Container Registry (CR) Gateways.

Generic Gateway is _not enabled by default_ and provides Brigade developers with the ability to trigger scripts based on messages received from any platform that can send a POST HTTP request.


## Intro to Generic Gateway 

Generic Gateway listens and accepts `POST` JSON messages at two different endpoints, `/simpleevents/v1/:projectID/:secret` and `/cloudevents/v02/:projectID/:secret`. When one of these endpoints is called, Brigade will respond by creating a Build with a `simpleevent` event or a `cloudevent` one, respectively. 

### SimpleEvent

Generic Gateway accepts valid JSON objects (thereafter called `SimpleEvent`) at the `/simpleevents/v1/:projectID/:secret` endpoint. Here's a sample JSON object:

```json
{
    "ref": "refs/heads/changes",
    "commit": "b60ad9543b2ddbbe73430dd6898b75883306cecc"
}
```

We are using `v1` in the endpoint path in case we want to have another version of `SimpleEvent` in the future.

### CloudEvent

Generic Gateway accepts [CloudEvents](https://cloudevents.io/) messages at the `/cloudevents/v02/:projectID/:secret` endpoint. As you can understand from the endpoint path, Generic Gateway currently supports [version 0.2](https://github.com/cloudevents/spec/blob/v0.2/spec.md) of the [CloudEvents specification](https://github.com/cloudevents/spec), using the [CloudEvents Go SDK](https://github.com/cloudevents/sdk-go). CloudEvent messages should be JSON encoded and transferred via HTTP(S).

## Generic Gateway

Generic Gateway is disabled by default, but can easily be turned on during installation or upgrade of Brigade:

```
$ helm install -n brigade brigade/brigade --set genericGateway.enabled=true
```

This will enable and configure the Generic Gateway Deployment/Services/RBAC permissions. However, be aware that Generic Gateway is not exposed outside the cluster by default. In case you are certain from a security perspective that you want to expose the Generic Gateway via a [Kubernetes LoadBalancer Service](https://kubernetes.io/docs/concepts/services-networking/#loadbalancer), you can install Brigade via the following command:

```
$ helm install -n brigade brigade/brigade --set genericGateway.enabled=true,genericGateway.service.type=LoadBalancer
```

Alternatively, for enhanced security, you can install an SSL proxy (like `cert-manager`) and direct it to the Generic Gateway Service.

## Using the Generic Gateway

As mentioned, Generic Gateway accepts POST requests at `/simpleevents/v1/:projectID/:secret` and `/cloudevents/v02/:projectID/:secret` endpoint. These requests should also carry a JSON payload (either a SimpleEvent or a CloudEvent).

Moreover:
- `projectID` is the Brigade Project ID
- `secret` is a custom secret for this specific project, in order to properly authenticate Generic Gateway requests. In other words, each project that wants to accept Generic Gateway events should have its own Generic Gateway secret. This secret serves as a simple authentication mechanism.

When you create a new Brigade Project via Brig CLI, you can optionally create such a secret by using the Advanced Options during `brig project create`. This secret must contain only alphanumeric characters. If you provide an empty string, Brig will generate and output a secret for you.

*Important*: If you do not go into "Advanced Options" during `brig project create`, a secret will not be created and you will not be able to use Generic Gateway for your project. However, you can always use `brig project create --replace` (or just `kubectl edit` your project Secret) to update your project and include a `genericGatewaySecret` string value.

### Calling the SimpleEvents endpoint

When calling the Generic Gateway `simpleevents` endpoint, you must include a SimpleEvent with a custom JSON payload such as:

```json
{
    "ref": "refs/heads/changes",
    "commit": "b60ad9543b2ddbbe73430dd6898b75883306cecc",
    "key1": "value1",
    "key2": "value2"
}
```

You can use curl to test it:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{
    "ref": "refs/heads/changes",
    "commit": "b60ad9543b2ddbbe73430dd6898b75883306cecc",
    "key1": "value1",
    "key2": "value2"
}' \
  http://localhost:8000/simpleevents/v1/PROJECT_ID/SECRET
```

If you do not wish to provide any payload, you can send empty POST data or just an empty JSON object (`{}`). 

---
**NOTE**

If you plan to use this type of event with source control, you should specify `ref` and `commit` values. These are used to configure the specific revision that Brigade will pull from your repository. If both values are missing, your Build's `ref` will be set to `master`. Of course, you can also provide any other values you need.

---

### Calling the CloudEvent endpoint

When calling the Generic Gateway `cloudevents/v02` endpoint, you must include a valid [0.2 CloudEvent](https://github.com/cloudevents/spec/blob/v0.2/spec.md) message such as:

```json
{
    "type":   "com.example.file.created",
    "source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
    "id":     "ea35b24ede421",
    "specversion": "0.2",
    "data": {
        "ref": "refs/heads/changes",
        "commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28",
        "key1": "value1",
        "key2": "value2"
  }
}
```

You can use curl to test it:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{ "type":   "com.example.file.created",
    "source": "/providers/Example.COM/storage/account#fileServices/default/{new-file}",
    "id":     "ea35b24ede421",
    "specversion": "0.2",
    "data": {
        "ref": "refs/heads/changes",
        "commit": "63c09efb6eb544f41a48901a6d0cc6ddfa4adb28",
        "key1": "value1",
        "key2": "value2"
  }}' \
  http://localhost:8000/cloudevents/v02/PROJECT_ID/SECRET
```

Bear in mind that 0.2 Cloud Events specification requires non-empty and proper values for "type","source","id" and "specversion" ([source](https://github.com/cloudevents/spec/blob/v0.2/spec.md#type)).

A CloudEvent may include domain-specific information in the `data` field ([source](https://github.com/cloudevents/spec/blob/v0.2/spec.md#data-attribute)). 

---
**NOTE**

If you plan to use this type of event with source control, you should specify `ref` and `commit` values in the `data` field in order to to configure the specific revision that Brigade will pull from your repository. If both values are missing, your Build's `ref` will be set to `master`.

---

## Sample Brigade.js

Here is a sample Brigade.js file that could be used as a base for your own scripts that respond to both Generic Gateway events. 

```javascript
const { events, Job } = require("brigadier");

events.on("simpleevent", (e, p) => {  // handler for a SimpleEvent
  var echo = new Job("echosimpleevent", "alpine:3.8");
  echo.tasks = [
    "echo Project " + p.name,
    "echo event type: $EVENT_TYPE"
  ];
  echo.env = {
    "EVENT_TYPE": e.type
  };
  echo.run();
});

events.on("cloudevent", (e, p) => { // handler for a CloudEvent
  var echo = new Job("echocloudevent", "alpine:3.8");
  echo.tasks = [
    "echo Project " + p.name,
    "echo event type: $EVENT_TYPE"
  ];
  echo.env = {
    "EVENT_TYPE": e.type
  };
  echo.run();
});
```
