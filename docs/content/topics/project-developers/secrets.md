---
title: Secret Management
description: How to use and manage secrets in Brigade
section: project-developers
weight: 3
aliases:
  - /secrets
  - /topics/secrets.md
  - /topics/project-developers/secrets.md
---

# Secret Management

Brigade provides tools for storing sensitive data outside of your Brigade
scripts, and then passing that information into the jobs that need them.

This is accomplished by associating secrets with a given project, which the
project's script can then access as needed. Secrets are persisted on the
substrate in the project namespace.

## Adding a Secret to Your Project

Imagine a case where we need to pass a sensitive piece of information to one of
the jobs in a project's Brigade script. For example, we might need to pass an
authentication token to a job that must authenticate to a remote service.

We first need to set the secret key and value on the project via `brig`. There
are two methods to do this, both provided on the `brig project secret set`
command:

 * Set a secret in-line:

  ```console
  $ brig project secret set --project my-project --set foo=bar
  ```

  * Setting a secret via a `secrets.yaml` file:

  ```console
  $ echo "foo: bar" >> ./secrets.yaml
  $ brig project secret set --project my-project --file ./secrets.yaml
  ```

## Accessing a Secret within a Brigade script

Within a Brigade script (`brigade.js` or `brigade.ts` file), we can access any
of the secrets defined on our project.

This first example demonstrates direct access of a secret defined on a project
in the event handler code outside of any Job:

```javascript
// THIS IS NOT SAFE! IT LEAKS YOUR SECRET INTO THE LOGS.
const { events } = require("@brigadecore/brigadier");

events.on("brigade.sh/cli", "exec", async event => {
  console.log("Project secret foo = " + event.project.secrets.foo);
});

events.process();
```

For passing project secrets into a Job, the best practice is to do so via the
Job's environment. Brigade inherently treats all Job environment variables as
sensitive by default, so any instance where their values would otherwise be
logged will be redacted.

Project secrets that are accessed directly in a Job, i.e. outside of its
environment, *will* be leaked, both in the Job Pod resource on the substrate
and via `brig` commands that consume this information, e.g.
`brig event get --id <event id>`.

Here is an example of passing a project secret into a Job:

```javascript
const { Job, events } = require("@brigadecore/brigadier");

events.on("brigade.sh/cli", "exec", async event => {
  let job = new Job("second-job", "debian:latest", event);
  job.primaryContainer.environment = {"FOO_SECRET": event.project.secrets.foo};
  job.primaryContainer.command = ["bash"];
  job.primaryContainer.arguments = ["-c", "echo $FOO_SECRET | wc -c"];

  await job.run();
});

events.process();
```

In this case, we retrieve the secret from the project, and we pass it into the
new Job as an environment variable. When the job executes, it can access the
`foo` secret value as `$FOO_SECRET`. In this trivial example, we echo the
secret value and confirm the expected character count via `wc`.

## Image pull secrets for Worker and Jobs

An [image pull secret] is used by the substrate (Kubernetes) to pull an OCI
image on which a Worker's or Job's container is based. This is only necessary
if the image exists in a private repository.

Any/all image pull secrets must be pre-created on the substrate by an
[Operator] prior to use in Brigade. The secret must be created in the namespace
dedicated to a project for use by that project's Worker/Jobs.

As a reminder, a project's namespace can be found under `kubernetes.namespace`
when inspecting a project via:

```console
$ brig project get --id <project name> -o yaml
```

Once created, the name of the image pull secret may be added to the Worker
configuration on the intended project. For example, if an image pull secret
named `privateregistrysecret` exists on the substrate in the namespace for
project `my-project`, it can be added to the project definition like so:

```yaml
apiVersion: brigade.sh/v2-beta
kind: Project
metadata:
  id: my-project
spec:
  workerTemplate:
    kubernetes:
      imagePullSecrets:
      - privateregistrysecret
```

Worker and Job containers for events handled by the project will then have
access to the image pull secret.

[image pull secret]: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/
[Operator]: /topics/operators

## FAQ

**Why don't all jobs automatically get access to all of the project's secrets?
Why do I have to pass them to the `Job.environment`?**

Brigade is designed to use off-the-shelf Docker images. In the examples above,
we used the `debian:latest` image straight from DockerHub. We wouldn't want to
just automatically pass all of our information straight into that container.
For starters, doing so might inadvertently override an existing environment
variable of the same name. More importantly, the data might get misused or
unintentionally exposed by the container.

So we err on the side of safety.

**Can I encrypt my secrets?**

Brigade uses [Kubernetes Secrets] for holding sensitive data. Kubernetes Secret
data is base64-encoded but not encrypted by default. Operators looking to
encrypt these resources should reference the [Kubernetes documentation]. Note
that support for this feature will depend on the distribution used (e.g. the
cloud provider offering or other self-hosted option.)

Another option for Brigade developers to access secret values is via a trusted
key store external to Kubernetes, such as HashiCorp's [Vault] or [Azure Key
Vault]. The scripting logic can then use the appropriate CLI or API calls to
fetch secrets in the event handler(s) that require them.

[Kubernetes Secrets]: https://kubernetes.io/docs/concepts/configuration/secret/
[Kubernetes documentation]: https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/
[Vault]: https://www.vaultproject.io/
[Azure Key Vault]: https://azure.microsoft.com/en-us/services/key-vault/