---
title: A Brigade Quickstart
description: A Brigade Quickstart.
section: intro
---

## Install Brigade

The easiest way to install Brigade into your Kubernetes cluster is to install it using [Helm](https://helm.sh/), the Kubernetes Package Manager.

```bash
# add Brigade chart repo
helm repo add brigade https://brigadecore.github.io/charts
# install Brigade
helm install -n brigade brigade/brigade

# if you want to activate Generic Gateway, you should use this command
# helm install -n brigade brigade/brigade --set genericGateway.enabled=true
```

You will now have Brigade installed. [Kashti](https://github.com/brigadecore/kashti), the dashboard for your Brigade pipelines, is also installed in the cluster.

## Install brig

Brig is the Brigade command line client. You can use `brig` to create/update/delete new brigade Projects, run Builds, etc. To get `brig`, navigate to the [Releases page](https://github.com/brigadecore/brigade/releases/) and then download the appropriate client for your platform. For example, if you're using Linux or WSL, you can get the 1.1.0 version in this way:

```bash
wget -O brig https://github.com/brigadecore/brigade/releases/download/v1.1.0/brig-linux-amd64
chmod +x brig
mv brig ~/bin
```

Alternatively, you can use [asdf-brig](https://github.com/Ibotta/asdf-brig) to install & manage multiple versions of `brig`.

We have two quickstarts for you to check:

- The first one creates a Project that will pull source from a Version Control System (VCS). This is usually used as a CI/CD pipeline.
- The second one creates a Project that has no dependency on a Version Control System and its Builds will be triggered via POST requests on Brigade's [Generic Gateway](https://docs.brigade.sh/topics/genericgateway/). Think of this approach as some JavaScript code that will do stuff with containers and be triggered by a POST message (either a plain JSON one or a [CloudEvent](https://cloudevents.io/)).

By the way, this does not mean that you cannot combine these scenarios (i.e. have your Builds triggered by POST requests and your source code be pulled and acted upon) or think of alternative ways to use Brigade!

## Using Brigade with a Version Control System

### Creating A New Project - using a Version Control System

To create a new project, use `brig project create` and answer the prompts. Feel free to modify or leave all options at their defaults (just press Enter on every interactive prompt).

```bash
brig project create
```

Output would be similar to this:
```
? VCS or no-VCS project? VCS
? Project name brigadecore/empty-testbed
? Full repository name github.com/brigadecore/empty-testbed
? Clone URL (https://github.com/your/repo.git) https://github.com/brigadecore/empty-testbed.git
? Add secrets? No
Auto-generated a Shared Secret: "novxKi64FKWyvU4EPZulyo0o"
? Configure GitHub Access? No
? Configure advanced options No
Project ID: brigade-4897c99315be5d2a2403ea33bdcb24f8116dc69613d5917d879d5f
```

Here we're using the name 'brigadecore/empty-testbed' for our project, which points to a test repo on https://github.com/brigadecore/empty-testbed. Of course, don't forget to give a proper name to your project, as well as set the 'Clone URL' correctly. If it's wrong, your subsequent Builds will fail! For documentation on project creation, check [here](https://docs.brigade.sh/topics/projects/).

Now we can view the newly created project:
```bash
brig project list
```

Output would be something like:
```
NAME                    ID                                                              REPO
myusername/myproject    brigade-4897c99315be5d2a2403ea33bdcb24f8116dc69613d5917d879d5f  github.com/myusername/myproject
```

You can also do a `kubectl get secret` to view the [Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/) that was created for this Project. Bear in mind that Brigade stores information about its entities (Project/Build) in Secrets.

```
NAME                                                             TYPE                                  DATA   AGE
brigade-4897c99315be5d2a2403ea33bdcb24f8116dc69613d5917d879d5f   brigade.sh/project                    24     2m
...other cluster Secrets...
```

### Creating Your First Brigade Build

Conceptually, a Brigade Build is a set of instructions (written in JavaScript, in a file called `brigade.js`) that will run upon a triggering of an event. Events can be something like a git push, a Docker push or just a webhook. In this way, you can think of Builds as jobs/tasks/pipelines.

Let's now create a new file called `brigade.js` with the following content:

```javascript
const { events, Job } = require("brigadier");
events.on("exec", () => {
  var job = new Job("do-nothing", "alpine:3.8");
  job.tasks = [
    "echo Hello",
    "echo World"
  ];

  job.run();
});
```

When the `exec` event is triggered, Brigade will create a Build based on this brigade.js file. This Build will create a single job that will start an image based on `alpine:3.8` which will simply do a couple of echoes. The Kubernetes Pod that will be created to run this Build job will have a name starting with 'do-nothing'. As you can probably guess, a new Kubernetes Secret will also be created to store information about this Build.

Moreover, you can check out [this tutorial](https://docs.brigade.sh/intro/) for more on creating scripts.

### Running a Build

To create and run a Brigade Build for the brigade.js file we wrote, we will use `brig`.

```bash
brig run brigadecore/empty-testbed -f brigade.js
```

With this command we are using the `brigade.js` file we just created. We could let Brigade use the `brigade.js` file in the `brigadecore/empty-testbed` repo, however this does not contain an `exec` event handler so nothing would happen. The usage of the custom `brigade.js` file we created lets us define a custom `exec` event handler.

So, this command will trigger the `exec` event in the `brigade.js` file we created and show the detailed output, which will be similar to this:

```
Event created. Waiting for worker pod named "brigade-worker-01d0y7bcxs6ke0yayrx6nbvm39".
Build: 01d0y7bcxs6ke0yayrx6nbvm39, Worker: brigade-worker-01d0y7bcxs6ke0yayrx6nbvm39
prestart: no dependencies file found
prestart: loading script from /etc/brigade/script
[brigade] brigade-worker version: 1.1.0
[brigade:k8s] Creating secret do-nothing-01d0y7bcxs6ke0yayrx6nbvm39
[brigade:k8s] Creating pod do-nothing-01d0y7bcxs6ke0yayrx6nbvm39
[brigade:k8s] Timeout set at 900000
[brigade:k8s] Pod not yet scheduled
[brigade:k8s] default/do-nothing-01d0y7bcxs6ke0yayrx6nbvm39 phase Pending
[brigade:k8s] default/do-nothing-01d0y7bcxs6ke0yayrx6nbvm39 phase Succeeded
done
```

As you can see, Brigade created a new Pod for this Build (called `do-nothing-01d0y7bcxs6ke0yayrx6nbvm39`) that executed the Job. Let's get its logs.

```bash
kubectl logs do-nothing-01d0y7bcxs6ke0yayrx6nbvm39
```

Output:
```
Hello
World
```

Moreover, you can get the details for this Build from `brig`.

```bash
brig build list
```

Output:
```
ID                              TYPE    PROVIDER        PROJECT                                                         STATUS          AGE
01d0y7bcxs6ke0yayrx6nbvm39      exec    brigade-cli     brigade-4897c99315be5d2a2403ea33bdcb24f8116dc69613d5917d879d5f  Succeeded       4m
```

What is not directly visible here is the fact that the Job Pod used [git-sidecar](https://github.com/brigadecore/brigade/tree/master/git-sidecar) as its [initContainer](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/). The `git-sidecar` container pulled the source code from the `master` branch of the `github.com/brigadecore/empty-testbed` repo and stored it in an `emptyDir` [Volume](https://kubernetes.io/docs/concepts/storage/volumes/). The Job Pod also mounts this Volume and therefore has access to the source code from the repo.

Brigade also supports private repos, you should select `Yes` on `Configure GitHub Access?` question of `brig project create` and fill out the prompts.
Last but not least, Brigade can listen to events (and trigger Builds) from a VCS via a [Gateway](https://docs.brigade.sh/topics/gateways/). We have some gateways for you to use, check them out: 

- [GitHub](https://docs.brigade.sh/topics/github/) gateway
- [GitLab](https://github.com/lukepatrick/brigade-gitlab-gateway) gateway
- [BitBucket](https://github.com/lukepatrick/brigade-bitbucket-gateway) gateway

## Using Brigade with Generic Gateway (no Version Control System)

### Creating Brigade.js

You should have activated Generic Gateway during Brigade installation using `helm install -n brigade brigade/brigade --set genericGateway.enabled=true`.

Since we are not going to use a repository that contains a `brigade.js` file for the Build to run, we should create one locally. So, first of all, create a `brigade.js` file with the following contents:

```javascript
const { events, Job } = require("brigadier");

events.on("simpleevent", (e, p) => {  // handler for a SimpleEvent
  var echo = new Job("echosimpleevent", "alpine:3.8");
  echo.tasks = [
    "echo Project " + p.name,
    "echo event type: $EVENT_TYPE",
    "echo payload " + JSON.stringify(e.payload)
  ];
  echo.env = {
    "EVENT_TYPE": e.type
  };
  echo.run();
});
```

On each Build, Brigade Worker will run this file and create a container with a name starting from `echosimpleevent` based on `alpine:3.8` image which will echo some details about the Project and the event itself.

### Creating a new Project

We will create a Project that will listen for a [SimpleEvent](https://docs.brigade.sh/topics/genericgateway/), which you can think of as a simple JSON object.

To create a new project, use `brig project create`. You should make sure to add a `brigade.js` script, either using the `Default script ConfigMap name` or the `Upload a default brigade.js script` option.

```
? VCS or no-VCS project? no-VCS
? Project Name brigadecore/empty-testbed
? Add secrets? No
? Secret for the Generic Gateway (alphanumeric characters only). Press Enter if you want it to be auto-generated mysecret
? Default script ConfigMap name
? Upload a default brigade.js script brigade.js
? Configure advanced options [? for help] (y/N)
Project ID: brigade-4897c99315be5d2a2403ea33bdcb24f8116dc69613d5917d879d5f
```

### Send a SimpleEvent to the Generic Gateway

Next, we should send a simple JSON object to test our Generic Gateway. By default, Generic Gateway is not exposed to the outside world, since it is created as a [Cluster IP Kubernetes Service](https://kubernetes.io/docs/concepts/services-networking/service/).

To send traffic to a Cluster IP Service, we can use `kubectl port-forward`. The following command will open a tunnel from local port 8081 to port 8081 on the Generic Gateway Service.

```bash
kubectl port-forward service/brigade-brigade-generic-gateway 8081:8081
```

To send a JSON message to Generic Gateway, open a new shell and try this curl command:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{
    "key1": "value1",
    "key2": "value2"
}' \
  http://localhost:8081/simpleevents/v1/brigade-4897c99315be5d2a2403ea33bdcb24f8116dc69613d5917d879d5f/mysecret
```

Pay attention to the Generic Gateway URL for a SimpleEvent, it is `HOST:PORT/simpleevents/v1/Project-ID/genericGatewaySecret`

If all is well, you'll get back a 200 result: 

```json
{"status":"Success. Build created"}
```

### View this Build's log

Let's now see what happened in our Worker Pods:

```bash
kubectl get pods
```

You'll see that two Pods were created for this Build. The worker Pod and the `echosimpleevent` Pod. For more details into Brigade design, check [here](https://docs.brigade.sh/topics/design/).

```
NAME                                               READY   STATUS      RESTARTS   AGE
brigade-brigade-api-6c5d6f4dcb-fzqg5               1/1     Running     0          22m
brigade-brigade-ctrl-6cc46c6769-mm2v7              1/1     Running     0          22m
brigade-brigade-generic-gateway-6f6496958f-sms5q   1/1     Running     0          2m51s
brigade-kashti-6bf64b8458-vhqxb                    1/1     Running     0          22m
brigade-worker-01d6tzxyqwbym5qdxa8wj4s6mx          0/1     Completed   0          111s
echosimpleevent-01d6tzxyqwbym5qdxa8wj4s6mx         0/1     Completed   0          16s
```

Let's get the `echosimpleevent` logs:

```bash
kubectl logs echosimpleevent-01d6tzxyqwbym5qdxa8wj4s6mx
```

You'll see the output that we requested in our Brigade.js file:

```
Project github.com/brigadecore/empty-testbed
event type: simpleevent
payload {\n    "key1": "value1",\n    "key2": "value2"\n}
```

To learn more about the Generic Gateway, check our docs [here](https://docs.brigade.sh/topics/genericgateway/).

## Kashti

You can also see the Project/Build output combination details in [Kashti](https://github.com/brigadecore/kashti). Kashti is by default visible only from within the cluster, so you need a `kubectl port-forward` from your local machine to the Kubernetes Service for Kashti.

```bash
kubectl port-forward service/brigade-kashti 8000:80
```

Then, you can navigate to `http://localhost:8000` to see Kashti dashboard with your Project and Build. Feel free to check [brigadeterm](https://github.com/slok/brigadeterm) which is similar to Kashti but runs inside your terminal.

## Vacuum

Brigade contains a utility (called `vacuum`) that runs as a Kubernetes CronJob and periodically (default: hourly) deletes Builds (i.e. corresponding Secrets and Pods). You can run `kubectl get cronjob` to get its details and possibly configure it.

## Cleanup

To remove created resources:

```bash
# delete project
brig project delete brigadecore/empty-testbed
# remove Brigade
helm delete brigade --purge
```