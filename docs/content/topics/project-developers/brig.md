---
title: The Brig CLI
description: Using the Brigade CLI
aliases:
  - /brig.md
  - /topics/brig.md
  - /topics/project-developers/brig.md
---

TODO: update per v2

# Brig: The Brigade CLI

Brig is a command line tool for interacting with Brigade. It allows Brigade
users to learn about their projects and builds, and provides a convenient way
to execute scripts.

## Basic Usage

### Creating a project

To create a project with brig, run `brig project create` and follow the interactive prompts,
supplying project name, GitHub repo details (if not derived from project name), optional secrets
and optional advanced configuration.

An example setup might look like the following:

```console
$ brig project create
? VCS or no-VCS project? VCS
? Project name brigadecore/empty-testbed
? Full repository name github.com/brigadecore/empty-testbed
? Clone URL (https://github.com/your/repo.git) https://github.com/brigadecore/empty-testbed.git
? Add secrets? No
Auto-generated a Shared Secret: "FweBxcwJvcbTTuW5CquyPtHM"
? Configure GitHub Access? No
? Configure advanced options No
```

To read more about project configuration, see [projects](../docs/topics/projects.md).

### Running scripts

Brig has built-in help text that you can access easily by adding `-h` or `--help`
to any command (e.g. `brig -h` or `brig project -h`). One of the most frequent
usages of Brig is to send a script to the server.

This example program sends a Brigade JavaScript file to a brigade server.

Example usage:

```console
$ brig run my-org/my-project
```

The above will load the local `./brigade.js` to Brigade and execute it within the project
`my-org/my-project`.

By default, Brig requests that the event `exec` be run. You can override this by
supplying a `--event=NAME` flag. For example, try executing the following script:

```javascript
const { events } = require('brigadier');

events.on('exec', () => {
  console.log('Hello from brig!');
});
```

A more complete example:

```console
$ brig run --file my/brigade.js --config my/brigade.json --namespace my-builds technosophos/myproject
```

The above looks for `./my/brigade.js` along with the `./my/brigade.json` configuration file and
sends them to the Brigade server inside of the Kubernetes `my-builds` namespace. It executes within the project
`technosophos/myproject`.

The output of the master process is written to STDOUT.

For more details on how the dependencies section of the `brigade.json` config file is used, see the [dependencies](dependencies.md) doc.

### Starting the Brigade web dashboard

Brig comes with a web dashboard, Kashti, which can be launched using the `brig dashboard` command.
The dashboard runs in your Kubernetes cluster, and Brig creates a tunnel, on `localhost`, then opens
the default browser. On top of the global settings of Brig (the Kubernetes context, config, and namespace),
the port of the local tunnel and disabling the launch of the default browser can be configured with flags.

```
$ brig dashboard --help
Flags:
  -h, --help             help for dashboard
      --open-dashboard   open the dashboard in the browser (default true)
      --port int         local port for the Kashti dashboard (default 8081)

$ brig dashboard
Connecting to kashti at http://localhost:8081...
Connected! When you are finished with this session, enter CTRL+C.
```

![Brigade web dashboard, Kashti](https://user-images.githubusercontent.com/686194/33646819-7d19d222-da06-11e7-8513-82e521fda608.gif)

### Starting the Brigade CLI dashboard

Brig also comes with a CLI dashboard that communicates directly with the Brigade API.
This dashboard is an integration of the community project [Brigadeterm](https://github.com/slok/brigadeterm/),
and can be launched using the `brig term` command. The refresh interval of the dashboard can be passed as a
flag to the `brig term` command.

![Brigade CLI dashboard](https://docs.brigade.sh/img/brig-term.png)

## How Brig Works

Brig uses your `$KUBECONFIG` to find out about your Kubernetes cluster. It then
authenticates directly to Kubernetes and interacts with Brigade and Kubernetes
APIs.
