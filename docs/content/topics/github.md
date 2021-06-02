---
title: GitHub Integration
description: How to provide GitHub integration for triggering Brigade builds from GitHub events.
aliases:
  - /github.md
  - /gateways/github
  - /gateways/github.md
  - /topics/gateways/github.md
  - /topics/design/github.md
  - /intro/topics/github.md
---

# GitHub Integration

Brigade can optionally provide GitHub integration for triggering Brigade builds from GitHub events
via the [Brigade Github App][brigade-github-app] project.  By default, this gateway is disabled.

To get set up, follow the [instructions][brigade-github-app-readme] to create and configure a GitHub App.
This App can then be used across one or more repositories, as opposed to the older, [OAuth approach](https://github.com/brigadecore/github-gateway-oauth) requiring configuration for each individual respository.

Next, to enable this gateway for a Brigade installation, set the `brigade-github-app.enabled` to `true`:

```
helm upgrade brigade2 ~/charts/brigade --namespace brigade2 --set brigade-github-app.enabled=true
kubectl rollout status deployment brigade2-apiserver -n brigade2 --timeout 1m
```

The rest of the `brigade-github-app` chart values can either be placed under the key of the same
name in the main values file for the Brigade chart (here called `brigade-values.yaml`), or they can be
placed in a separate yaml file.  If the latter, be sure all of the configuration is still under this
sub-chart's name, like this:

```
$ cat brigade-github-app-values.yaml
brigade-github-app:
  enabled: true
  # Set this to true to enable Kubernetes RBAC support (recommended)
  rbac:
    enabled: true

  # Image configuration
  registry: brigadecore
  name: brigade-github-app
...
  github:
    # The x509 PEM-formatted keyfile GitHub issued for your App.
    key: |
      -----BEGIN RSA PRIVATE KEY-----
...
      -----END RSA PRIVATE KEY-----
    checkSuiteOnPR: true
    appID: <appID>
...

$ helm install -n brigade/brigade -f brigade-values.yaml -f brigade-github-app-values.yaml
```

To link this GitHub App up with GitHub repositories by way of Brigade projects, continue following the
[README.md](https://github.com/brigadecore/brigade-github-app/blob/master/README.md#6-add-brigade-projects-for-each-github-project).

[brigade-github-app]: https://github.com/brigadecore/brigade-github-app
[brigade-github-app-readme]: https://github.com/brigadecore/brigade-github-app/blob/master/README.md