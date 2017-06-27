# Working with Projects

In the previous section we installed Acid using Helm. In this section, we will
configure a new Acid project, and have GitHub push and pull request events trigger
Acid builds.

## Creating an Acid Project

The Acid server tracks separate configuration for each project you set up. And
to create and manage these configurations, we use a simple YAML file and a Helm
chart.

### A Basic Project YAML

First, let's create a new project and point it to the GitHub project https://github.com/technosophos/zolver

We'll create the simplest config file possible:

my-project.yaml:
```yaml
project: "technosophos/zolver"
repository: "github.com/technosophos/zolver"
# Used by GitHub to compute hooks.
# TODO: MAKE SURE YOU CHANGE THIS. It's basically a password.
secret: "MySuperSecret"
```

Now that this is done, let's install this project into Acid. Recall that in the
last section we used `helm` to install Acid. We also use it to install projects.

### Installing a Project Chart

```console
$ helm install acid/acid-project -n my-project -f my-project.yaml
```

Note that `-n my-project` provides a name for the project, which you will be able
to see with `helm ls`. And `-f my-project.yaml` loads the YAML file you wrote
above.

Your project configuration can now be managed by Helm. Use `helm upgrade` to change
the configuration. And `helm delete` will remove the project.

## Configuring GitHub

_FIXME: We could probably add some screenshots here_

We want to build our project each time a new commit is pushed to master, and each
time we get a new Pull Request.

To do this, log into your project on https://github.com.


From your project...

1. Click on the settings menu item (with the gear icon)
2. In the left navigation, click on `Webhooks`
3. On the Webhooks screen, click `Add Webhook`
4. Complete the form:
  - Payload URL should be the URL to your ACID server
  - Content-Type should be `application/json`
  - Secret should be your secret in the YAML file (`MySuperSecret`)
  - In the radio buttons, choose `Let me select individual events`
    - Select `push` and `pull request`
  - Make sure the `Activate` checkbox is selected
  - Click the `Add Webhook` button

The next time you push to the repository, the webhook system should trigger a build.
