// ============================================================================
// NOTE: This is the actual acid.js file for testing the Acid project.
// Be careful when editing!
// ============================================================================

// This is a Go project, so we want to set it up for Go.
gopath = "/go"

// To set GOPATH correctly, we have to override the default
// path that Acid sets.
localPath = gopath + "/src/github.com/" + pushRecord.repository.full_name;

// Create a new job
job1 = new Job("acid-test");

// Since this is Go, we want a go runner.
job1.image = "technosophos/acid-go:latest";

// Set a few environment variables.
job1.env = {
    "DEST_PATH": localPath,
    "GOPATH": gopath
};

// Run three tasks in order.
job1.tasks = [
  "go get github.com/Masterminds/glide",
  "glide install",
  "make test-unit"
];

// Run and wait for it to finish.
job1.run(pushRecord);
