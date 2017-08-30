// ============================================================================
// NOTE: This is the actual acid.js file for testing the Acid project.
// Be careful when editing!
// ============================================================================
const { events, Job, Group} = require("libacid")

function build(e, project) {
  // This is a Go project, so we want to set it up for Go.
  var gopath = "/go"

  // To set GOPATH correctly, we have to override the default
  // path that Acid sets.
  var localPath = gopath + "/src/github.com/" + project.repo.name;


  // Create a new job to run Go tests
  var goBuild = new Job("acid-test", "golang:1.8");

  // Set a few environment variables.
  goBuild.env = {
      "DEST_PATH": localPath,
      "GOPATH": gopath
  };

  // Run Go unit tests
  goBuild.tasks = [
    "go get github.com/Masterminds/glide",
    // Need to move the source into GOPATH so vendor/ works as desired.
    "mkdir -p " + localPath,
    "mv /src/* " + localPath,
    "cd " + localPath,
    "glide install --strip-vendor",
    "make test-unit"
  ];

  // Run the acid worker tests
  var jsTest = new Job("acid-js-build", "node:8");
  jsTest.tasks = [
    "cd /src/acid-worker",
    "yarn install",
    "yarn test"
  ];

  // Run in parallel
  Group.runAll([jsTest, goBuild])
}

events.on("push", build)
events.on("pull_request", build)
