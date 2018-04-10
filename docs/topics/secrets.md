# Secret Management

Brigade provides tools for storing sensitive data outside of your `brigade.js` scripts,
and then passing that information into the jobs that need them. Brigade accomplishes
this by making use of Kubernetes secrets.

## Adding a Secret to Your Project

Imagine a case where we need to pass a sensitive piece of information to one of the jobs
in our `brigade.js`. For example, we might need to pass an authentication token to a job that
must authenticate to a remote service.

We do this by storing the secret inside of the project definition.

Project definitions are typically managed by Helm. As you may recall from the installation
manual, a new project is created like this:

```console
$ helm install brigade/brigade-project -n my-project -f my-values.yaml
```

The `my-values.yaml` file looks something like this:

```yaml
project: "deis/empty-testbed"
repository: "github.com/deis/empty-testbed"
cloneURL: "https://github.com/deis/empty-testbed.git"
sharedSecret: "aaaaaaaaaaaaaa"
namespace: "default"
secrets: {}
```

Note the empty `secrets` object at the end. That is the place for you to put your own
secrets.

For example, we can add a database password like this:

```yaml
project: "deis/empty-testbed"
repository: "github.com/deis/empty-testbed"
cloneURL: "https://github.com/deis/empty-testbed.git"
sharedSecret: "aaaaaaaaaaaaaa"
namespace: "default"
secrets:
  dbPassword: supersecret
```

As usual, you can use `helm upgrade` to update the values. Everything in the `secrets`
section will be stored inside of your project's Kubernetes Secret.

```console
$ helm upgrade my-project brigade/brigade-project -f my-values.yaml
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
trusted key store such as Vault.

Alternatively, you could use `secretKeyRef` to reference existing secrets already in your
Kubernetes cluster.

**I don't want to use Helm to manage my project/secrets. Can I do it manually?**

Yes. Helm is there to make your life easier, but you can manage project secrets manually.
You cannot, however, make manual modifications and then go back to using Helm. Doing so
may result in lost data.
