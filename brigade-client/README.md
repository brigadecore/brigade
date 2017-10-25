# Lightweight Script Deployer

This example program sends a Brigade JavaScript file to a brigade server.

Example usage:

```console
$ lsd my-org/my-project
```

The above will load the local `./brigade.js` to Brigade and execute it within the project
`my-org/my-project`.

A more complete example:

```console
$ lsd --file my/brigade.js --ns my-builds technosophos/myproject
```

The above looks for `./my/brigade.js` and sends it to the Brigade server inside of
the Kubernetes `my-builds` namespace. It executes within the project
`technosophos/myproject`.

The output of the master process is written to STDOUT.
