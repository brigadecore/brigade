# Brigade: Event-based Scripting for Kubernetes

![Build Status](http://badges.technosophos.me/v1/github/build/Azure/brigade/badge.svg?branch=master)

Script simple and complex workflows using JavaScript. Chain together containers,
running them in parallel or serially. Fire scripts based on times, GitHub events,
Docker pushes, or any other trigger. Brigade is the tool for creating pipelines
for Kubernetes.

- JavaScript scripting
- Project-based management
- Configurable event hooks
- Easy construction of pipelines
- Check out the [docs](https://docs.brigade.sh/) to get started.

 <!-- [![asciicast](https://asciinema.org/a/JBsjOpah4nTBvjqDT5dAWvefG.png)](https://asciinema.org/a/JBsjOpah4nTBvjqDT5dAWvefG) -->

## The Brigade Technology Stack

- Brigade :heart: JavaScript: Writing Brigade pipelines is as easy as writing a few lines of JavaScript.
- Brigade :heart: Kubernetes: Brigade is Kubernetes-native. Your builds are translated into
  pods, secrets, and services
- Brigade :heart: Docker: No need for special plugins or elaborate extensions. Brigade uses
  off-the-shelf Docker images to run your jobs. And Brigade also supports DockerHub
  webhooks.
- Brigade :heart: GitHub: Brigade comes with built-in support for GitHub, DockerHub, and
  other popular web services. And it can be easily extended to support your own
  services.

The [design introduction](https://docs.brigade.sh/topics/design/) introduces Brigade concepts and
architecture.

## Quickstart

### Install Brigade

The easiest way to install Brigade into your Kubernetes cluster is to install it using [Helm](https://helm.sh/), the Kubernetes Package Manager.

```bash
# add Brigade chart repo
helm repo add brigade https://brigadecore.github.io/charts
# install Brigade - this also installs Kashti
helm install -n brigade brigade/brigade
```

You will now have Brigade installed. [Kashti](https://github.com/brigadecore/kashti), the dashboard for your Brigade pipelines, is also installed in the cluster.

### Install brig

Brig is the Brigade command line client. You can use `brig` to create/update/delete new brigade Projects, run Builds, etc. To get `brig`, navigate to the [Releases page](https://github.com/brigadecore/brigade/releases/) and then download the appropriate client for your platform. For example, if you're using Linux or WSL, you can get the 0.20.0 version in this way:

```bash
wget -O brig https://github.com/brigadecore/brigade/releases/download/v0.20.0/brig-linux-amd64
chmod +x brig
mv brig ~/bin
```

Alternatively, you can use [asdf-brig](https://github.com/Ibotta/asdf-brig) to install & manage multiple versions of `brig`.

### Creating A New Project

To create a new project, use `brig project create` and answer the prompts. Feel free to modify or leave all options at their defaults (just press Enter on every interactive prompt).

```bash
brig project create
```

Output would be similar to this:
```
? Project name deis/empty-testbed
? Full repository name github.com/deis/empty-testbed
? Clone URL (https://github.com/your/repo.git) https://github.com/deis/empty-testbed.git
? Add secrets? No
Auto-generated a Shared Secret: "novxKi64FKWyvU4EPZulyo0o"
? Configure GitHub Access? No
? Configure advanced options No
Project ID: brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
```

Here we're using the name 'deis/empty-testbed' for our project, which points to a test repo on 'https://github.com/deis/empty-testbed'. Of course, don't forget to give a proper name to your project, as well as set the 'Clone URL' correctly. If it's wrong, your subsequent Builds will fail! For documentation on project creation, check [here](https://docs.brigade.sh/topics/projects/).

Now we can view the newly created project:
```bash
brig project list
```

Output would be something like:
```
NAME                    ID                                                              REPO
myusername/myproject    brigade-2e9bb93bf149536a951d236772ae8be77a3cef9335c82bf39fc18c  github.com/myusername/myproject
```

You can also do a `kubectl get secret` to view the [Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/) that was created for this Project. Bear in mind that Brigade stores information about its entities (Project/Build) in Secrets.

```
NAME                                                             TYPE                                  DATA   AGE
brigade-2e9bb93bf149536a951d236772ae8be77a3cef9335c82bf39fc18c   brigade.sh/project                    24     2m
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
brig run deis/empty-testbed -f brigade.js
```

This will trigger the `exec` event and show you the detailed output, which will be similar to this:

```
Event created. Waiting for worker pod named "brigade-worker-01d0y7bcxs6ke0yayrx6nbvm39".
Build: 01d0y7bcxs6ke0yayrx6nbvm39, Worker: brigade-worker-01d0y7bcxs6ke0yayrx6nbvm39
prestart: no dependencies file found
prestart: loading script from /etc/brigade/script
[brigade] brigade-worker version: 0.20.0
[brigade:k8s] Creating secret do-nothing-01d0y7bcxs6ke0yayrx6nbvm39
[brigade:k8s] Creating pod do-nothing-01d0y7bcxs6ke0yayrx6nbvm39
[brigade:k8s] Timeout set at 900000
[brigade:k8s] Pod not yet scheduled
[brigade:k8s] default/do-nothing-01d0y7bcxs6ke0yayrx6nbvm39 phase Pending
[brigade:k8s] default/do-nothing-01d0y7bcxs6ke0yayrx6nbvm39 phase Succeeded
done
```

As you can see, Brigade created a new pod for this Build (called `do-nothing-01d0y7bcxs6ke0yayrx6nbvm39`) that executed our job. Let's get its logs.

```bash
kubectl logs do-nothing-01d0y7bcxs6ke0yayrx6nbvm39
```

Output:
```
Hello
world
```

Moreover, you can get the details for this Build from `brig`.

```bash
brig build list
```

Output:
```
ID                              TYPE    PROVIDER        PROJECT                                                         STATUS          AGE
01d0y7bcxs6ke0yayrx6nbvm39      exec    brigade-cli     brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac  Succeeded       4m
```

You can also see this Project/Build output combination in Kashti. Kashti is by default visible only from within the cluster, so you need a `kubectl port-forward` from your local machine to the Kubernetes Service for Kashti.

```bash
kubectl port-forward service/brigade-kashti 8000:80
```

Then, you can navigate to `http://localhost:8000` to see Kashti dashboard with your Project and Build.

Brigade contains a utility (called `vacuum`) that runs as a Kubernetes CronJob and periodically (default: hourly) deletes Builds (i.e. corresponding Secrets and Pods). You can run `kubectl get cronjob` to get its details and possible configure it.

### Cleanup

To remove created resources:

```bash
# delete project
brig  project delete deis/empty-testbed
# remove Brigade
helm delete brigade --purge
```

## Related Projects

- [Kashti](https://github.com/brigadecore/kashti) - a dashboard for your Brigade pipelines.
- [Brigadeterm](https://github.com/slok/brigadeterm) - a simple terminal ui for brigade pipelining system.
- [Brigade exporter](https://github.com/slok/brigade-exporter) - a [Prometheus](https://prometheus.io) exporter to gather metrics from Brigade.
- Gateways
  - [BitBucket events](https://github.com/lukepatrick/brigade-bitbucket-gateway): Gateway Support for BitBucket repositories
  - [GitLab events](https://github.com/lukepatrick/brigade-gitlab-gateway): Gateway Support for GitLab repositories
  - [Kubernetes events](https://github.com/brigadecore/brigade-k8s-gateway): Gateway that listens to Kubernetes event stream
  - [Event Grid gateway](https://github.com/radu-matei/brigade-eventgrid-gateway): Gateway for Azure Event Grid events
  - [Cron Gateway](https://github.com/technosophos/brigade-cron): Schedule events to run at a particular time
  - [Trello and Generic Webhooks](https://github.com/technosophos/brigade-trello): Experimental gateway for Trello and for generic webhooks
  - [Draft Pack for Building Custom Gateways](https://github.com/technosophos/draft-brigade): Build your own gateway [in 5 minutes](http://technosophos.com/2018/04/23/building-brigade-gateways-the-easy-way.html)
  - [Azure DevOps / VSTS gateway](https://github.com/radu-matei/brigade-vsts-gateway): Gateway for Azure DevOps / VSTS events

## Brigade :heart: Developers

To get started head to the [developer's guide](https://docs.brigade.sh/topics/developers/)

Brigade is well-tested on Minikube and [Azure Kubernetes Service](https://docs.microsoft.com/en-us/azure/aks/).

# Contributing

This project welcomes contributions and suggestions. Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

# Support & Feedback

We have a slack channel! [Kubernetes/#brigade](https://kubernetes.slack.com/messages/C87MF1RFD) Feel free to join for any support questions or feedback, we are happy to help. To report an issue or to request a feature open an issue [here](https://github.com/brigadecore/brigade/issues)

[brigade-project-chart]: https://github.com/brigadecore/brigade-charts/tree/master/charts/brigade-project
