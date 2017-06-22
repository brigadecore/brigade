// ============================================================================
// NOTE: This is the actual acid.js file for testing the Acid project.
// Be careful when editing!
// ============================================================================
/* global Job WaitGroup events */

// This handles a Push webhook.
events.push = function(e) {
  // This is a Go project, so we want to set it up for Go.
  var gopath = "/go"

  // To set GOPATH correctly, we have to override the default
  // path that Acid sets.
  var localPath = gopath + "/src/github.com/" + e.request.repository.full_name;


  // Create a new job
  var goBuild = new Job("acid-test");

  // Since this is Go, we want a go runner.
  goBuild.image = "golang:1.8";

  // Set a few environment variables.
  goBuild.env = {
      "DEST_PATH": localPath,
      "GOPATH": gopath
  };

  // Run three tasks in order.
  goBuild.tasks = [
    "date",
    "echo Begin test-unit",
    "go get github.com/Masterminds/glide",
    "go get github.com/jteeuwen/go-bindata/...",
    // Need to move the source into GOPATH so vendor/ works as desired.
    "mkdir -p " + localPath,
    "mv /src/* " + localPath,
    "cd " + localPath,
    "glide install --strip-vendor",
    "make test-unit"
  ];

  var jsLint = new Job("acid-js-build");

  jsLint.image = "technosophos/acid-node:latest";
  jsLint.tasks = [
    "date",
    "echo Begin test-js",
    "cd /src"
    "npm install -g --quiet eslint",
    "make test-js"
  ];

  // Run both jobs in parallel, and wait for then both to finish.
  var waiter = new WaitGroup()

  waiter.add(jsLint)
  waiter.add(goBuild)
  waiter.run()
}
