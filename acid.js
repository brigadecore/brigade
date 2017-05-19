// ============================================================================
// NOTE: This is the actual acid.js file for testing the Acid project.
// Be careful when editing!
// ============================================================================
/* global Job WaitGroup pushRecord */

// This is a Go project, so we want to set it up for Go.
var gopath = "/go"

// To set GOPATH correctly, we have to override the default
// path that Acid sets.
var localPath = gopath + "/src/github.com/" + pushRecord.repository.full_name;

// Create a new job
var goBuild = new Job("acid-test");

// Since this is Go, we want a go runner.
goBuild.image = "technosophos/acid-go:latest";

// Set a few environment variables.
goBuild.env = {
    "DEST_PATH": localPath,
    "GOPATH": gopath
};

// Run three tasks in order.
goBuild.tasks = [
  "go get github.com/Masterminds/glide",
  "glide install",
  "make test-unit"
];

var jsLint = new Job("acid-js-build");

jsLint.image = "technosophos/acid-node:latest";
jsLint.tasks = [
  "npm install -g --quiet eslint",
  "make test-js"
];

// Run and wait for it to finish.
goBuild.background();
jsLint.background();

// Wait for both jobs to finish
var waiter = new WaitGroup()

waiter.add(jsLint)
waiter.add(goBuild)
waiter.wait()
