---
title: 'Tutorial 3: Projects & Events'
description: 'Writing your first CI pipeline, Part 3'
section: intro
---

# Writing your first CI pipeline, Part 3

This tutorial begins where [Tutorial 2][part2] left off. We’ll walk through the process for configuring your newly created Github repository with Brigade for testing new features. We'll configure a new Brigade project, and have Github push events to trigger Brigade builds.

## Create a Brigade project

The Brigade server tracks separate configuration for each project you set up. To create and manage these configurations, we use the [brig](https://github.com/Azure/brigade/tree/master/brig) cli.

Here we create a project for our GitHub repo:

```console
 $ brig project create
? Project Name bacongobbler/uuid-generator
? Full repository name github.com/bacongobbler/uuid-generator
? Clone URL (https://github.com/your/repo.git) https://github.com/bacongobbler/uuid-generator.git
? Add secrets? No
Auto-generated a Shared Secret: "mDXUDZyDsTUHw4KZIMPOQMN1"
? Configure GitHub Access? No
? Configure advanced options No
Project ID: brigade-5ea0b3d7707afb5d04d55544485da6aff4f58006c1633f4ae0cb11
```

Note: to explore the advanced options, each prompt offers further info when `?` is entered.

## Configuring Github

We want to build our project each time a new commit is pushed to master, and each time we get a new Pull Request.

To do this, follow the [Brigade GitHub App][brigade-github-app] documentation to set up
a GitHub App.  During configuration, copy the shared secret above (`mDXUDZyDsTUHw4KZIMPOQMN1`) and set this as the
"Webhook secret" value for the App.

![GithHub Webhooks](https://docs.brigade.sh/img/img3.png)

We'll need to upgrade our Brigade server with our `brigade-github-app` sub-chart configuration filled in:

```console
$ helm inspect values brigade/brigade > brigade-values.yaml
$ # Add configuration under the `brigade-github-app` section
$ helm upgrade brigade brigade/brigade -f brigade-values.yaml
```

We can then get the IP needed to update the "Webhook URL" entry for our App.  Run this command on your
Kubernetes cluster, and look for the `brigade-github-app` line:

```console
$ kubectl get service
NAME                                TYPE           CLUSTER-IP     EXTERNAL-IP    PORT(S)        AGE
brigade-server-brigade-api          ClusterIP      10.0.34.228    <none>         7745/TCP       1d
brigade-server-brigade-github-app   LoadBalancer   10.0.69.248    10.21.77.9     80:31980/TCP   1d
```

You will use the `EXTERNAL-IP` address to form the full URL: `http://10.21.77.9/events/github`.
Update the App's "Webhook URL" with this value.  (Note: it is preferred that DNS be set up instead of
a hard-coded IP.  See the [Brigade GitHub App docs][brigade-github-app] for more.)

The next time you push to the repository, the webhook system should trigger a build.

> For more on configuring GitHub, see [the GitHub Guide](../topics/github.md)

After configuring Brigade to test new features, read [part 4 of this tutorial][part4] to write a new feature to the uuid-generator project, which will trigger a test build using Brigade.

[part2]: ../tutorial02
[part4]: ../tutorial04
[brigade-github-app]: https://github.com/Azure/brigade-github-app