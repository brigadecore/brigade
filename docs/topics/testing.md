# Testing Brigade Scripts

Now that we've written our `brigade.js` scripts, we're ready to confirm that the
Javascript is correct and functioning as intended.  Here we demonstrate the use
of a few utilities we can employ to do so.

## Javascript testing with brigtest

[brigtest](https://github.com/technosophos/brigtest) is a testing tool designed
to vet the Javascript portion of a `brigade.js` script without actually launching
any containers or requiring a Kubernetes cluster.  It can optionally mock
events, Jobs and projects.

Follow the [installation instructions](https://github.com/technosophos/brigtest#installing)
to set `brigtest` up on your machine.  For these examples, we'll assume `brigtest`
is installed and available globally.

Next, let's create an example `brigade.js` file:

```javascript
const { events, Job } = require("brigadier");

events.on("exec", (e, p) => {
  var one = new Job("one", "alpine:3.4");
  var two = new Job("two", "alpine:3.4");

  one.tasks = ["echo world"];
  one.run().then( result => {
    two.tasks = ["echo hello " + result.toString()];
    two.run().then( result2 => {
      console.log(result2.toString())
    });
  });
})
```

Now, let's run it through `brigtest` passing `-x` to check syntax:

```console
$ brigtest -f brigade.js -x
✨  Done in 0.22s.
```

This is a great way to first check that the Javascript is properly formatted.

To test our script with a mocked event, we can just drop the `-x` and brigtest
will use the default `exec` event if not otherwise specified via `-e <event>`:

```console
$ brigtest -f brigade.js
done
✨  Done in 0.24s.
```

Since brigtest doesn't actually launch the Jobs specified in our script, we don't
see their intended output.  To mock these Job outputs and/or supply a non-default
config, check out brigtest's [README.md](https://github.com/technosophos/brigtest/blob/master/README.md#modeling-behavior)

At this stage, we can be confident that our Javascript is correct and our use of
the `brigadier` library is as well.  So, we're all set to integration test.

## Integration testing with Brigade Integration Test

[Brigade Project Integration Test](https://github.com/blimmer/brigade-project-integration-test)
is a great project that lays the foundation for a way to integration test a Brigade project
on a provided Kubernetes cluster.  It is currently designed to work with
[Minikube](https://github.com/kubernetes/minikube).

To quickly get started with the project, check out the
[README.md](https://github.com/blimmer/brigade-project-integration-test/blob/master/README.md).

Once your kube context/`kubectl` is pointing to a minikube cluster, you can run the default
tests as instructed.  This will install a default Brigade server, a default Brigade project
and confirm all is in working order by issuing a `brig run` against the Brigade project,
using the `brigade.js` that lives at the repo's root.

We now have a great start for supplying a non-default Brigade project and `brigade.js`
script for testing.  Although supporting these overrides is still to be natively supported
in the tool (as of writing), it's simply a matter of adjusting a few strings and re-running
the test script to cover our example `brigade.js` file above.

In fact, to cover our script above, we can modify the main [bats](https://github.com/sstephenson/bats)
[script](https://github.com/blimmer/brigade-project-integration-test/blob/master/test/tests.bats) in
this project with an assertion around output.

As we expect to see `"hello world"` in the console output after the pipeline finishes,
our new test can look like the following:


```bash
@test "run should output 'hello world'" {
  run $BRIG_RUN
  assert_output --partial "hello world"
}
```
