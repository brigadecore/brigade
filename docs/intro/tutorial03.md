# Writing your first CI pipeline, Part 3

This tutorial begins where [Tutorial 2][part2] left off. We’ll walk through the process for configuring your newly created Github repository with Acid for testing new features. We'll configure a new Acid project, and have Github push events to trigger Acid builds.

## Create an Acid project

The Acid server tracks separate configuration for each project you set up. And to create and manage these configurations, we use a simple YAML file and a Helm chart.

First, let's create a new project and point it to the GitHub project we just created, *uuid-generator*

We'll create the simplest config file possible. Substituting *bacongobbler* for your own Github username, open the file `uuid-generator.yaml` and write in the following:

```yaml
project: "bacongobbler/uuid-generator"
repository: "github.com/bacongobbler/uuid-generator"
cloneURL: "https://github.com/bacongobbler/uuid-generator.git"
# Used by GitHub to compute hooks.
# MAKE SURE YOU CHANGE THIS. It's basically a password.
sharedSecret: "MySuperSecret"
# Use this to have Acid update your project about the build.
# You probably want this if you want pull requests or commits to show
# the build status.
github:
  token: "github oauth token"
```

Make sure to **not** commit this to source control. It contains private data that should not be publicized in a git repository.

To use a Github OAuth token so your Pull Request statuses are updated...

1. Go to https://github.com/settings/tokens/new and enter your password if prompted
2. Give the token a description, such as `acid project: uuid-generator`
3. Grant the token full *repo* scope so Acid can update Pull Request statuses

<img src="img/img3.png" style="height: 500px;" />

4. Click *Generate token*
5. Copy the personal access token in the next screen and add it to `uuid-generator.yaml`

### Install the project chart

Now that we have written the project chart, let's install this project into Acid. Recall that in the [Quick install guide](install.md) we used `helm` to install Acid. We also use it to install projects.

```
$ helm install acid/acid-project --name uuid-generator -f uuid-generator.yaml
```

Note that `-n uuid-generator` provides a name for the project, which you will be able to see with `helm ls`. And `-f uuid-generator.yaml` loads the YAML file you wrote above.

Your project configuration can now be managed by Helm. Use `helm upgrade` to change the configuration. And `helm delete` will remove the project. See `man helm-upgrade` for more options and information regarding these commands.

## Configuring Github

We want to build our project each time a new commit is pushed to master, and each time we get a new Pull Request.

To do this, log into your project (substituting *bacongobbler* for your own Github username) on https://github.com/bacongobbler/uuid-generator/settings.

From your project...

1. In the left navigation, click on `Webhooks`
2. On the Webhooks screen, click `Add Webhook`
3. Complete the form:
  - Payload URL should be the URL to your ACID server
  - Content-Type should be `application/json`
  - Secret should be your secret in the YAML file (`MySuperSecret`)
  - In the radio buttons, choose `Let me select individual events`
    - Select `push` and `pull request`
  - Make sure the `Activate` checkbox is selected
  - Click the `Add Webhook` button

<img src="img/img4.png" style="height: 500px;" />

The next time you push to the repository, the webhook system should trigger a build.

After configuring Acid to test new features, read [part 4 of this tutorial][part4] to write a new feature to the uuid-generator project, which will trigger a test build using Acid.


[part2]: tutorial02.md
[part4]: tutorial04.md
