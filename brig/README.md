# Brig: The Brigade CLI

Brig is a command line tool for interacting with Brigade. It allows Brigade
users to learn about their projects and builds, and provides a convenient way
to execute scripts.

## Basic Usage

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
const { events } = require("brigadier")

events.on("exec", () => {
  console.log("Hello from brig!")
})
```

A more complete example:

```console
$ brig run --file my/brigade.js --namespace my-builds technosophos/myproject
```

The above looks for `./my/brigade.js` and sends it to the Brigade server inside of
the Kubernetes `my-builds` namespace. It executes within the project
`technosophos/myproject`.

The output of the master process is written to STDOUT.

## Building Brig

To build Brig, clone the [Brigade repository](https://github.com/Azure/brigade)
to `$GOPATH/src/github.com/Azure/brigade` and then run `make bootstrap brig`.

If you have $GOPATH issues, you may need to [add the Brigade binary](https://github.com/Azure/brigade/issues/447) to your path.

## How Brig Works

Brig uses your `$KUBECONFIG` to find out about your Kubernetes cluster. It then
authenticates directly to Kubernetes and interacts with Brigade and Kubernetes
APIs.
