---
title: A Brigade Quickstart
description: A Brigade Quickstart.
section: intro
weight: 22
---

In this QuickStart, you will install Brigade, create a project and execute it.

* [Prerequisites](#prerequisites)
* [Install Brigade](#install-brigade)
* [Log in to Brigade](#log-in-to-brigade)
* [Create a Project](#create-a-project)
* [Trigger an Event](#trigger-an-event)

## Prerequisites

* [A development Kubernetes cluster](#create-a-cluster).
* [Brigade CLI](#install-the-brigade-cli) installed.
* [Helm] CLI v3+ installed.
* [kubectl] CLI installed.
* Free disk space. The installation requires sufficient free disk space and will fail if your disk is nearly full.

> Please take note that the default configuration is not secure and is not appropriate for any shared cluster.
> This configuration is appropriate for evaluating Brigade on a local development cluster, and should not be used in production.

### Create a Cluster

If you do not already have a development cluster, we recommend using [KinD].
KinD runs a Kubernetes cluster locally using [Docker].
[Minikube] also works well for local development.

1. Install [KinD]. See the KinD documentation for full installation instructions, below are instructions for common environments:

    **linux**
    ```bash
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
    chmod +x ./kind
    mv ./kind /usr/local/bin
    ```

    **macos with Homebrew**
    ```bash
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-darwin-amd64
    chmod +x ./kind
    mv ./kind /usr/local/bin
    ```

    **windows**
    ```powershell
    mkdir -force $env:USERPROFILE\bin
    (New-Object Net.WebClient).DownloadFile("https://kind.sigs.k8s.io/dl/v0.11.1/kind-windows-amd64", "$ENV:USERPROFILE\bin\kind.exe")
    $env:PATH+=";$env:USERPROFILE\bin"
    ```

    The script above downloads kind.exe and adds it to your PATH for the current session.
    Add the following line to your [PowerShell Profile](https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/) to make the change permanent.

    ```powershell
    $env:PATH+=";$env:USERPROFILE\bin"
    ```

1. Create a Kubernetes cluster by running the following command:
    ```
    kind create cluster
    ```

1. Verify that you can connect to the cluster using kubectl:
    ```
    kubectl cluster-info
    ```

[Helm]: https://helm.sh/docs/intro/install/
[Minikube]: https://minikube.sigs.k8s.io/docs/start/
[KinD]: https://kind.sigs.k8s.io/docs/user/quick-start/
[kubectl]: https://kubernetes.io/docs/tasks/tools/#kubectl
[Docker]: https://docs.docker.com/get-docker/

### Install the Brigade CLI

Install the Brigade CLI, brig, by copying the appropriate binary from our releases page into a directory on your machine that is included in your PATH environment variable.

**linux**
```bash
curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.1/brig-linux-amd64
chmod +x /usr/local/bin/brig
```

**macos**
```bash
curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.1/brig-darwin-amd64
chmod +x /usr/local/bin/brig
```

**windows**
```powershell
mkdir -force $env:USERPROFILE\bin
(New-Object Net.WebClient).DownloadFile("https://github.com/brigadecore/brigade/releases/download/v2.0.0-beta.1/brig-windows-amd64.exe", "$ENV:USERPROFILE\bin\brig.exe")
$env:PATH+=";$env:USERPROFILE\bin"
```

The script above downloads brig.exe and adds it to your PATH for the current session.
Add the following line to your [PowerShell Profile](https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/) to make the change permanent.

```powershell
$env:PATH+=";$env:USERPROFILE\bin"
```

## Install Brigade

Install Brigade on your local development cluster. See our [Installation] instructions for full instructions suitable for production clusters.

1. Enable Helm's experimental OCI support by setting the `HELM_EXPERIMENTAL_OCI` environment variable to 1.

    **posix**
    ```bash
    export HELM_EXPERIMENTAL_OCI=1
    ```

    **powershell**
    ```powershell
    $env:HELM_EXPERIMENTAL_OCI=1
    ```

1. Create a directory to store the Brigade Helm charts.

    **posix**
    ```bash
    mkdir -p ~/charts
    ```

    **powershell**
    ```powershell
    mkdir -force $env:USERPROFILE/charts
    ```

1. Run the following commands to install Brigade.

    ```
    helm chart pull ghcr.io/brigadecore/brigade:v2.0.0-beta.1
    helm chart export ghcr.io/brigadecore/brigade:v2.0.0-beta.1 -d ~/charts
    helm install brigade2 ~/charts/brigade --namespace brigade2 --create-namespace
    kubectl rollout status deployment brigade2-apiserver -n brigade2 --timeout 5m
    ```
    
    Wait for the Brigade deployment to be ready.
    If the deployment fails, proceed to the [installation troubleshooting](/intro/install/#troubleshooting) section.

Now that Brigade is running, you need to determine the address of the Brigade API so that you can use it later in this QuickStart:

### Port Forward a Local Cluster

If you are running a cluster locally, use port forwarding to make the Brigade API available via localhost:

**posix**

```
kubectl --namespace brigade2 port-forward service/brigade2-apiserver 8443:443 &>/dev/null &
```

**powershell**

```
& kubectl --namespace brigade2 port-forward service/brigade2-apiserver 8443:443 *> $null  
```

### Get External IP of a Remote Cluster

If you are running a cluster remotely, such as on a cloud provider, the Brigade API is available at the External IP of the brigade2-apiserver service:

```
kubectl get service --namespace brigade2 brigade2-apiserver -o=jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

[Installation]: /intro/install/

## Log in to Brigade

Authenticate to Brigade as the root user using demo password `F00Bar!!!`. The \--insecure flag instructs Brigade to ignore the self-signed certificate used by our local installation of Brigade.

**local clusters**

```
brig login --insecure --server https://localhost:8443 --root
```

If the address https://localhost:8443 does not resolve, double-check that the brigade2-apiserver service was successfully forwarded from the previous section.

**remote clusters**

Replace `IP_ADDRESS` with the External IP address of your cluster:

```
brig login --insecure --server https://IP_ADDRESS --root
```

## Create a Project

A Brigade [project] defines event handlers, such as the definition of a CI pipeline.
In this example project, the handler prints a message using a string passed in the event payload.

1. Download the example project to the current directory.

    **posix**
    ```bash
    curl -o project.yaml https://raw.githubusercontent.com/brigadecore/brigade/v2/examples/12-first-payload/project.yaml
    ```

    **powershell**
    ```bash
    (New-Object Net.WebClient).DownloadFile("https://raw.githubusercontent.com/brigadecore/brigade/v2/examples/12-first-payload/project.yaml", "$pwd\project.yaml")
    ```
1. Open project.yaml

    <script src="https://gist-it.appspot.com/https://raw.githubusercontent.com/brigadecore/brigade/v2/examples/12-first-payload/project.yaml"></script>

    The project defines a handler for the "exec" event, that reads the event payload string and prints it out with "Hello, PAYLOAD!".

1. Create the project in Brigade with the following command.

    ```
    brig project create --file project.yaml
    ```

1. List the defined projects with `brig project list` and verify that you see your new project:

    ```console
    $ brig project list
    ID           	DESCRIPTION                         	AGE
    first-payload	Demonstrates using the event payload	49m
    ```

[project]: /topics/projects/#an-introduction-to-projects

## Trigger an Event

With our project defined, you are now ready to trigger an event and watch your handler execute.

```
brig event create --project first-payload --payload Dolly --follow
```

Below is example output of a successful event handler:
```
Created event "7a5234d6-e2aa-402f-acb9-c620dfc20003".

Waiting for event's worker to be RUNNING...
2021-05-26T18:12:34.604Z INFO: brigade-worker version: v2.0.0-beta.1
2021-05-26T18:12:34.609Z DEBUG: writing default brigade.js to /var/vcs/.brigade/brigade.js
2021-05-26T18:12:34.609Z DEBUG: using npm as the package manager
2021-05-26T18:12:34.610Z DEBUG: path /var/vcs/.brigade/node_modules/@brigadecore does not exist; creating it
2021-05-26T18:12:34.610Z DEBUG: polyfilling @brigadecore/brigadier with /var/brigade-worker/brigadier-polyfill
2021-05-26T18:12:34.610Z DEBUG: found nothing to compile
2021-05-26T18:12:34.611Z DEBUG: running node brigade.js
Hello, Dolly!
```

## Cleanup

If you want to keep your Brigade installation, run the following command to remove the example project created in this QuickStart:

```
brig project delete first-payload
```

Otherwise, you can remove ALL resources created in this QuickStart by either:

* Deleting the KinD cluster that you created at the beginning with `kind delete cluster --name kind-kind` OR
* Preserving the cluster and uninstalling Brigade with `helm delete brigade2 -n brigade2`

## Next Steps

You now know how to install Brigade on a local development cluster, define a project, and trigger an event for the project.
Next learn how to [install and configure Brigade](/intro/install/) on a production cluster, or continue learning about
Brigade with our [CI pipeline tutorial](/intro/tutorial01/).

## Troubleshooting

* [Brigade installation does not finish successfully](/intro/install/#troubleshooting)
* [Login command hangs](#login-command-hangs)

### Login command hangs

If the brig login command hangs, check that you included the -k flag.
This flag is required because our local development installation of Brigade is using a self-signed certificate.


<!--
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
[brigade] brigade-worker version: 1.2.1
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

You can also see the Project/Build output combination details in [Kashti](https://github.com/brigadecore/kashti). Kashti is by default visible only from within the cluster, and `brig` has a helper command to create a port forwarding session from the Kashti service on Kubernetes to your local machine:

```bash
brig dashboard
Connecting to kashti at http://localhost:8081...
Connected! When you are finished with this session, enter CTRL+C.
```

Then, you can navigate to `http://localhost:8081` to see Kashti dashboard with your Project and Build. You can also use `brig term` to see a terminal dashboard.

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
-->