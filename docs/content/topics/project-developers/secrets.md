---
title: Secret Management
description: How to use and manage secrets in Brigade
aliases:
  - /secrets.md
  - /topics/secrets.md
  - /topics/project-developers/secrets.md
---

TODO: update per v2

# Secret Management

Brigade provides tools for storing sensitive data outside of your `brigade.js` scripts,
and then passing that information into the jobs that need them. Brigade accomplishes
this by making use of Kubernetes secrets.

## Adding a Secret to Your Project

Imagine a case where we need to pass a sensitive piece of information to one of the jobs
in our `brigade.js`. For example, we might need to pass an authentication token to a job that
must authenticate to a remote service.

We do this by storing the secret inside of the project definition.

Project definitions are typically managed by `brig`. As you may recall from the [installation
manual](../intro/install.md), a new project is created like via `brig project create`.

During the creation process, we are able to add secrets when prompted. If we'd forgotten to
add a secret and/or additional secrets need to be added to this existing project, we can do so via
`brig project create --replace -p <existing project name>`.

Here we add a new secret with key `dbPassword` and value `supersecret` to our existing
`brigadecore/empty-testbed` project:

```console
$ brig project create --replace -p brigadecore/empty-testbed
? Existing Project Name brigadecore/empty-testbed
? Full repository name github.com/brigadecore/empty-testbed
? Clone URL (https://github.com/your/repo.git) https://github.com/brigadecore/empty-testbed.git
? Add secrets? Yes
? 	Secret 1 dbPassword
? 	Value supersecret
? ===> Add another? No
Auto-generated a Shared Secret: "9qQvuplBx39r04Wd9EmyxjGA"
? Configure GitHub Access? No
? Configure advanced options No
Project ID: brigade-830c16d4aaf6f5490937ad719afd8490a5bcbef064d397411043ac
```

## Accessing a Secret within `brigade.js`

Within the `brigade.js` file, we can access any of the secrets defined on our project.

```javascript
// THIS IS NOT SAFE! IT LEAKS YOUR SECRET INTO THE LOGS.
const {events, Job, Group} = require("brigadier")

events.on("push", function(e, project) {
  console.log("My DB password is " + project.secrets.dbPassword)
})
```

Secrets can be selectively passed to jobs by setting environment variables.

```javascript
// THIS IS NOT SAFE! IT LEAKS YOUR SECRET INTO THE LOGS.
const {events, Job, Group} = require("brigadier")

events.on("push", function(e, project) {
  var j1 = new Job("secrets", "alpine:3.4")

  // Send the password to an environment variable.
  j1.env = {
    "DB_PASSWORD": project.secrets.dbPassword
  }

  // Print the env var to the log.
  j1.tasks = [
    "echo $DB_PASSWORD"
  ]

  j1.run()
}
```

In this case, we retrieve the secret from the project, and we pass it into the new
Job as an environment variable. When the job executes, it can access the `dbPassword`
as `$DB_PASSWORD`.

Note that behind the scenes, Brigade is storing the environment variables in another
Job-specific secret.

## FAQ

**Why don't all jobs get access to all of the secrets? Why do I have to pass them
to the `Job.env`?**

Brigade is designed to use off-the-shelf Docker images. In the examples above, we used the
`alpine:3.4` image straight from DockerHub. We wouldn't want to just automatically pass
all of our information straight into that container. For starters, doing so might
inadvertently override an existing environment variable of the same name. More
importantly, the data might get misused or unintentionally exposed by the container.

So we err on the side of safety.

**Can I encrypt my secrets?**

We use Kubernetes Secrets for holding sensitive data. As encrypted Secrets are
adopted into Kubernetes, we plan to support them. However, the present stable
version of Kubernetes Secrets only Base64-encodes data.

Our present recommendation is for Brigade developers to fetch the secret directly from a
trusted key store such as Vault.  See the [example](#fetching-secrets-from-a-remote-store) below.

Alternatively, you could use `secretKeyRef` to reference existing secrets already in your
Kubernetes cluster.  See the [example](#using-kubernetes-secrets) below.

**I don't want to use Helm to manage my project/secrets. Can I do it manually?**

Yes. Helm is there to make your life easier, but you can manage project secrets manually.
You cannot, however, make manual modifications and then go back to using Helm. Doing so
may result in lost data.

## Examples

---

## Fetching secrets from a remote store

Here is an example where we fetch secrets from a remote store
([Azure Key Vault](https://azure.microsoft.com/en-us/services/key-vault/))
for use by one or more Jobs.

To do so, we first run a Job to authenticate with the remote store, fetch all needed secrets
and save these secret values to a shared directory.  Then, we run a Job that consumes these values.

Here's what our `brigade.js` file looks like:

```javascript
const { events, Job, Group } = require("brigadier");

const sharedMountPrefix = `/mnt/brigade/share`;
const secrets = [
  "foo",
  "bar"
]

events.on("exec", (event, project) => {
  // Create Job to fetch secrets
  secretfetcher = new Job("secretfetcher", "microsoft/azure-cli:latest");
  secretfetcher.storage.enabled = true;

  // Login as an Azure Service Principal
  secretfetcher.tasks.push(
    `az login --service-principal \
    -u ${project.secrets.spID} \
    -p '${project.secrets.spPW}' \
    --tenant ${project.secrets.spTenant}`)

  // Fetch all secrets from the Azure Key Vault
  for (i in secrets) {
    secretfetcher.tasks.push(
      `az keyvault secret show --vault-name ${project.secrets.keyvault} \
      -n ${secrets[i]} --query value > ${sharedMountPrefix}/${secrets[i]}`);
  }

  // Create Job to consume secrets
  secretconsumer = new Job("secretconsumer", "alpine");
  secretconsumer.storage.enabled = true;

  // Consume all secrets
  for (i in secrets) {
    secretconsumer.tasks.push(
      `ls -haltr ${sharedMountPrefix}/${secrets[i]}`);
  }

  // Run jobs sequentially
  Group.runEach([secretfetcher, secretconsumer]);
})
```

To run this example, we replace `BRIGADE_PROJECT_NAME` with our Brigade project name and run:

```console
brig run BRIGADE_PROJECT_NAME -f brigade.js
```

When we check the logs of the `secretconsumer` job pod, we see:

```console
 $ kubectl logs secretconsumer-01d23ch2n6d1847ys6x48x1rez
-rwxrwxrwx    1 root     root           6 Jan 25 20:52 /mnt/brigade/share/foo
-rwxrwxrwx    1 root     root           6 Jan 25 20:52 /mnt/brigade/share/bar
```

## Using Kubernetes secrets

Kubernetes secrets can also be used for consuming secret data.

For this example, we first create the Kubernetes secret:

```console
kubectl create secret generic mysecret --from-literal=hello=world
```

This Kubernetes Secret has a name of `mysecret` and a single key-value pair as its data (key: hello, value: world).

Then, we create a `brigade.js` file with the following contents:

```javascript
const { events, Job } = require("brigadier");

events.on("exec", () => {
  var job = new Job("echo", "alpine");

  job.env = {
    mySecretReference: {
      secretKeyRef: {
        name: "mysecret",
        key: "hello"
      }
    }
  };

  job.tasks = [
    "echo hello ${mySecretReference}"
  ];

  job.run();
});
```

To recap, `mySecretReference` is the name of the variable in the job's environment, `secretKeyRef.name` is the name
of the Kubernetes secret, and `secretKeyRef.key` is the key pointing to the actual secret value (`world` in this example.)

As a result, `mySecretReference` points directly to this secret value in this job's environment.

To run this example, we replace `BRIGADE_PROJECT_NAME` with our Brigade project name and run:

```console
brig run BRIGADE_PROJECT_NAME -f brigade.js
```

We will then see `"hello world"` displayed in the logs.
