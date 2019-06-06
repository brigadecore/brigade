---
title: "Tutorial: Preview Environments"
description: Automate Creation of Development Preview Environments
section: intro
---

In this tutorial we'll create a workflow for creating Development Preview Environments.

Preview Environment is a dedicated Kubernetes Namespace where all of your applications 
and their dependencies are deployed. We will create a pipeline for both creation of these ephemeral environments, 
as well as auto-deployment of the applications themselves.

Preview Environments are a great place for early experimentation where changes to your 
application under development can be tested against production version of applications you're depending on.

## Setup

For the purpose of this tutorial we're assuming you have a Kubernetes cluster up and running and Brigade components are deployed to `brigade` namespace within it.  

We'll be relying on GitHub webhooks so make sure that `brigade-github-app.enabled` is set to `true` when installing `brigade` helm chart. You can learn more about GitHub integration [here](../../topics/github).  

[Docker for Desktop's Kubernetes](https://docs.docker.com/docker-for-mac/kubernetes/) 
cluster is sufficient to perform all steps from this tutorial.  

To accept incoming GitHub Webhooks, ensure your [ingress](../../topics/ingress/) is configured. Alternatively, on a desktop cluster, you can use a free version of the [Ngrok](https://ngrok.com/) service to establish secure tunneling. Follow this [excellent guide](https://stefanprodan.com/2018/expose-kubernetes-services-over-http-with-ngrok/) to set this up.

Brigade's [brig](../../intro/install/#brig) cli utility should be present on your machine.

## Git Repositories

We will use two GitHub repositories:

- [brigade-tutorial-config](https://github.com/brigadecore/brigade-tutorial-config): containing orchestration responsible for managing Brigade projects and creating new environments.
- [brigade-tutorial-app](https://github.com/brigadecore/brigade-tutorial-app): example microservice used to demonstrate release process.

GitHub token will be used in this tutorial. To generate it follow [this article](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line).

## Config Brigade Project

Let's create a new Brigade Project for our `brigade-tutorial-config` repository.

```console
$ brig project create --namespace brigade
? Project Name brigadecore/brigade-tutorial-config
? Full repository name github.com/brigadecore/brigade-tutorial-config
? Clone URL https://github.com/brigadecore/brigade-tutorial-config
```

## Create Preview Environment Pipeline

New environments will be created by executing a brigade script via `brig` cli. 

![Create Environment Pipeline](https://docs.brigade.sh/img/preview-environments-create.png)

## Create Namespace

Let's start by adding a new `brigade.js` script to the root of your repository. For now it will be responsible for creation of a new Kubernetes namespace.

```js

// brigade.js

const { events } = require("@brigadecore/brigadier");
const kubernetes = require("@kubernetes/client-node");

const kubeConfig = new kubernetes.KubeConfig();
kubeConfig.loadFromDefault();

const k8sCoreClient = kubeConfig.makeApiClient(kubernetes.Core_v1Api);

const protectedEnvironment = namespaceName => {
  const protectedNamespaces = ["default", "kube-public", "kube-system", "brigade"];

  if (protectedNamespaces.includes(namespaceName)) {
    return true;
  }
  return false;
};

const createNamespace = async namespaceName => {
  const existingNamespace = await k8sCoreClient.listNamespace(
    true,
    "",
    `metadata.name=${namespaceName}`,
  );

  if (existingNamespace.body.items.length) {
    console.log(`Namespace "${namespaceName}" already exists`);
    return;
  }

  const namespace = new kubernetes.V1Namespace();
  namespace.metadata = new kubernetes.V1ObjectMeta();
  namespace.metadata.name = namespaceName;

  await k8sCoreClient.createNamespace(namespace);
};

const provisionEnvironment = async (environmentName, projects) => {
  await createNamespace(environmentName);
};

events.on("exec", event => {
    const payload = JSON.parse(event.payload);
    const { name } = payload;

    if (!name) {
      throw Error("Environment name must be specified");
    }
    if (protectedEnvironment(name)) {
      throw Error(`Environment '${name}' is protected`);
    }
    provisionEnvironment(name, projects).catch(error => {
      throw error;
    });
});

```

In the same directory add `payload.json` file with the following content:

```json
{"name": "bob"}
```

Let's run our workflow with `brig`:

```console
$ brig run brigadecore/brigade-tutorial-config -f brigade.js -p payload.json --namespace brigade
```

Command above will trigger a Brigade workflow by executing the `brigade.js` script with the data from `payload.json`.

At this point our `brigade.js` script will simply use the Kubernetes API to create a new namespace. The name of new namespace is set in payload json.

First, we compare the namespace with a list of `protected` namespace names and then we're using 
API to check if the namespace already exists, and if not, we create it. 

> **With the power of the Kubernetes API we can easily orchestrate any aspects of our pipeline.**

## Environment Dependencies

In this step we'll automate adding external dependencies like PostgreSQL to our new environment.

```js
const k8sAppClient = kubeConfig.makeApiClient(kubernetes.Apps_v1Api);

const deployDependencies = async environmentName => {
  const postgresqlStatefulSet = await k8sAppClient.listNamespacedStatefulSet(
    environmentName,
    undefined,
    undefined,
    undefined,
    undefined,
    "app=postgresql",
  );
  if (postgresqlStatefulSet.body.items.length) {
    console.log("postgresql already deployed");
  } else {
    const postgresql = new Job("postgresql", "lachlanevenson/k8s-helm:v2.12.3");
    postgresql.storage.enabled = false;
    postgresql.imageForcePull = true;
    postgresql.tasks = [
      `helm init --client-only && \
      helm repo update && \
      helm upgrade ${environmentName}-postgresql stable/postgresql \
      --install --namespace=${environmentName} \
      --set fullnameOverride=postgresql \
      --set postgresqlDatabase=products \
      --set resources.requests.cpu=50m \
      --set resources.requests.memory=156Mi \
      --set readinessProbe.initialDelaySeconds=60 \
      --set livenessProbe.initialDelaySeconds=60;`,
    ];
    await postgresql.run();
  }
};

const provisionEnvironment = async (environmentName, projects) => {
  await deployDependencies(environmentName);
};

```

Before deploying our PostgreSQL StatefulSet we check if one already exists. 
To list StatefulSets in a Namespace we need to use `listNamespacedStatefulSet` 
api endpoint that lives in `kubernetes.Apps_v1Api` set of APIs. 

Once we verify our dependency needs to be installed we create a `postgresql` Job that will run the `lachlanevenson/k8s-helm:v2.12.3` image as its worker. 
Lachlan Evenson has been building and publishing [docker images](https://hub.docker.com/r/lachlanevenson/k8s-helm/) with every version of Helm.

After the Job successfully completes, we can verify our PostgreSQL is installed:

```console
$ helm list

NAME            	STATUS  	CHART                	NAMESPACE
bob-postgresql  	DEPLOYED	postgresql-3.16.     	bob
```

## Environment ConfigMap


The Environment ConfigMap is created in the `brigade` namespace where we keep track of all projects we would like to be installed in our environment. One ConfigMap is created per environment.

Sample of data structure:

```yaml
projects:
  products:
    tag: prod
  orders:
    tag: prod
```

Our ConfigMap will be labeled with a `type: preview-environment-config` label. This label will be used as a selector by the application's `brigade.js` script at the time of a release.

```js
const createEnvironmentConfigMap = async (name, projects) => {
  const configMap = new kubernetes.V1ConfigMap();
  const metadata = new kubernetes.V1ObjectMeta();
  metadata.name = `preview-environment-${name}`;
  metadata.namespace = 'brigade';
  metadata.labels = {
    type: "preview-environment-config",
    environmentName: name,
  };
  configMap.metadata = metadata;
  configMap.data = {
    projects: yaml.dump(projects),
  };

  await k8sCoreClient.createNamespacedConfigMap('brigade', configMap);
};

const provisionEnvironment = async (environmentName, projects) => {
  await createEnvironmentConfigMap(environmentName, projects);
};

```

You can see the final implementation of the script in the reference repository:

https://github.com/brigadecore/brigade-tutorial-config/blob/master/brigade.js

## Products Project

`Products Service` is a sample application that will be deployed to our preview environment every time it is released (tagged with a `prod` tag). The instance of PostgreSQL deployed  during creation of the environment is a dependency of our service.

![Create Environment Pipeline](https://docs.brigade.sh/img/preview-environments-release.png)

## Service Brigade Project

Let's create a new Brigade Project for our `brigade-tutorial-app` repository. 

```console
$ brig project create --namespace brigade
? Project Name brigadecore/brigade-tutorial-app
? Full repository name github.com/brigadecore/brigade-tutorial-app
? Clone URL https://github.com/brigadecore/brigade-tutorial-app
Auto-generated a Shared Secret: "uSEtlJicRK3RhRWiOatImwBs"
? Configure GitHub Access? Yes
? OAuth2 token <my-github-token>
```

## GitHub Webhook

To enable auto-deployment of our service to all (interested) preview environments we first need to enable a GitHub Webhook that will notify Brigade of a new release.

Note: if using docker-for-desktop Kubernetes cluster follow [this guide](https://stefanprodan.com/2018/expose-kubernetes-services-over-http-with-ngrok/) to set up Ngrok tunneling to your local `brigade-github-app` service.

- In your GitHub repository go to `Settings -> Webhooks -> Add Webhook`
- In `Payload URL` enter your exposed brigade url e.g. http://e8432c17.ngrok.io/events/github
- Change `Content type` to `application/json`
- In `Secret` enter the shared secret that was auto-generated during run of `brig project create` command above.
- Choose `Let me select individual events`. Unselect all options that have been preselected and choose `Branch or tag creation`. For the purpose of this tutorial we will be interested in tags created events only.

## Service Implementation

Our Python Products Microservice will expose a REST Api that will handle adding products to a database.

Service codebase has 3 main parts:

**Service Code:**  
`product` folder contains actual service code with product model defined in `models.py` 
and service Api in `service.py`

**Database Migrations:**  
`migrations` folder contains Alembic database migrations scripts which will be executed 
on each release.

**Helm Charts:**  
`charts` folder contains Helm chart for our service. Brigade script will use it to push 
latest version of the service to our Preview Environments.

To review full implementation details head over to the service repository: https://github.com/brigadecore/brigade-tutorial-app

## Brigade Script

When executing service deployment script we'll be fetching relevant git commit sha 
for `prod` tag by calling GitHub REST API. To do so, we'll need to install additional 
`node-fetch` dependency. We can do so by adding `brigade.json` file alongside our brigade 
script. Brigade's `prestart` hook will load this file and install any dependencies listed there:

```json
{
  "dependencies": {
    "node-fetch": "^2.3.0"
  }
}
```

In brigade script implementation we will handle `create` events coming from 
GtiHub gateway upon tag creation.  

Script flow is as follows:

- Grab a payload from incoming event

```js
events.on("create", event => {
  const payload = JSON.parse(event.payload);
  if (payload.ref_type !== "tag") {
    console.log("skipping, not a tag commit");
    return;
  }
  deployToEnvironments(payload).catch(error => {
    throw error;
  });
});
```
- Use the Kubernetes API to fetch all ConfigMaps from the `brigade` namespace of type `preview-environment-config`
- For every preview environment ConfigMap:
    - find out if that environment is interested in our service and 
    - check if current payload tag matches tag specified for the environment.

```js
const deployToEnvironments = async payload => {
  const tag = payload.ref;
  const environmentConfigMaps = await k8sClient.listNamespacedConfigMap(
    BRIGADE_NAMESPACE,
    true,
    undefined,
    undefined,
    undefined,
    "type=preview-environment-config",
  );
  if (!environmentConfigMaps.body.items.length) {
    throw Error("No environment configMaps found");
  }
  for (const configMap of environmentConfigMaps.body.items) {
    const projects = yaml.safeLoad(configMap.data.projects);
    const config = projects[PROJECT_NAME];
    if (config && config.tag === tag) {
      const { environmentName } = configMap.metadata.labels;
      const gitSha = await getTagCommit(tag, config.org, config.repo);
      await deploy(environmentName, gitSha);
    }
  }
};
```
- Our docker images are tagged with git sha so we're calling GitHub API to get it for 
  current tag.

```js
const getTagCommit = async (tag, org, repo) => {
  console.log(`getting commit sha for tag ${tag}`);
  const tagUrl = `${GITHUB_API_URL}/${org}/${repo}/git/refs/tags/${tag}`;
  const response = await fetch(tagUrl, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      Authorization: `token ${process.env.BRIGADE_REPO_AUTH_TOKEN}`,
    },
  });
  if (response.ok) {
    const commit = await response.json();
    return commit.object.sha;
  }
  throw Error(await response.text());
};
```

- We execute a Brigade Job that runs Helm deployment by providing appropriate docker image tag

```js
const deploy = async (environmentName, gitSha) => {
  console.log("deploying helm charts");
  const service = new Job(
    "brigade-tutorial-app",
    "lachlanevenson/k8s-helm:v2.12.3",
  );
  service.storage.enabled = false;
  service.imageForcePull = true;
  service.tasks = [
    "cd /src",
    `helm upgrade ${environmentName}-products \
    charts/products --install \
    --namespace=${environmentName} \
    --set image.tag=${gitSha} \
    --set replicaCount=1`,
  ];
  await service.run();
};
```

You can see the final implementation of the script in the reference repository:

https://github.com/brigadecore/brigade-tutorial-app/blob/master/brigade.js

## Release

To release our application we follow these steps:

- Commit our changes to GitHub
- Build and push Docker image
- Add a `prod` tag to the head of our repository to trigger release

Once `prod` tag is added, GitHub will send payload event to our Brigade GitHub Gateway which will start our release pipeline.

## Summary

In this tutorial we've learnt how to orchestrate creation of isolated Preview Environments where changes can be pushed continuously after every code commit. You can use a similar approach for releases to your long lived environments, like staging or production. Head over to [brigade-tutorial-config](https://github.com/brigadecore/brigade-tutorial-config) and [brigade-tutorial-app](https://github.com/brigadecore/brigade-tutorial-app) repositories to see complete implementation of brigade scripts used in this tutorial.
