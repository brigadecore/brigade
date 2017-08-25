# Lightweight Script Deployer

This example program sends an Acid JavaScript file to an acid server.

Example usage:

```console
$ lsd my-org/my-project
```

The above will load the local `./acid.js` to Acid and execute it within the project
`my-org/my-project`.

A more complete example:

```console
$ lsd --file my/acid.js --ns my-builds technosophos/myproject
```

The above looks for `./my/acid.js` and sends it to the Acid server inside of
the Kubernetes `my-builds` namespace. It executes within the project
`technosophos/myproject`.

The output of the master process is written to STDOUT.
