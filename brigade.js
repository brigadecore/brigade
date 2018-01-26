// ============================================================================
// NOTE: This is the actual brigade.js file for testing the Brigade project.
// Be careful when editing!
// ============================================================================
const { events, Job, Group} = require("brigadier")

function build(e, project) {
  // This is a Go project, so we want to set it up for Go.
  var gopath = "/go"

  // To set GOPATH correctly, we have to override the default
  // path that Brigade sets.
  var localPath = gopath + "/src/github.com/" + project.repo.name;


  // Create a new job to run Go tests
  var goBuild = new Job("brigade-test", "golang:1.9");

  // Set a few environment variables.
  goBuild.env = {
      "DEST_PATH": localPath,
      "GOPATH": gopath
  };

  // Run Go unit tests
  goBuild.tasks = [
    "go get github.com/golang/dep/cmd/dep",
    // Need to move the source into GOPATH so vendor/ works as desired.
    "mkdir -p " + localPath,
    "mv /src/* " + localPath,
    "cd " + localPath,
    "dep ensure",
    "make test-unit"
  ];

  // Run the brigade worker tests
  var jsTest = new Job("brigade-js-build", "node:8");
  jsTest.tasks = [
    "cd /src/brigade-worker",
    "yarn install",
    "yarn test"
  ];

  // Run in parallel
  Group.runAll([jsTest, goBuild])
}

events.on("push", build)
events.on("pull_request", build)
events.on("image_push", (e, p) => {
  console.log(e.payload)
  var m = "New image pushed"

  if (project.secrets.SLACK_WEBHOOK) {
    var slack = new Job("slack-notify")

    slack.image = "technosophos/slack-notify:latest"
    slack.env = {
      SLACK_WEBHOOK: project.secrets.SLACK_WEBHOOK,
      SLACK_USERNAME: "BrigadeBot",
      SLACK_TITLE: "DockerHub Image",
      SLACK_MESSAGE: m + " <https://" + project.repo.name + ">",
      SLACK_COLOR: "#00ff00"
    }

    slack.tasks = ["/slack-notify"]
    slack.run()
  } else {
    console.log(m)
  }
})

events.on("functional_test", (e, p) => {
  var gopath = "/go"

  // To set GOPATH correctly, we have to override the default
  // path that Brigade sets.
  var localPath = gopath + "/src/github.com/" + p.repo.name;


  // Create a new job to run Go tests
  var goBuild = new Job("brigade-test", "golang:1.9");

  const j = new Job("functional-test", "golang:1.9")
  j.env = {
    GOPATH: "/src"
  }

  // Set a few environment variables.
  j.env = {
      "DEST_PATH": localPath,
      "GOPATH": gopath
  };

  // Run Go unit tests
  j.tasks = [
    "go get github.com/golang/dep/cmd/dep",
    "mkdir -p " + localPath,
    "mv /src/* " + localPath,
    "cd " + localPath,
    "dep ensure",
	  "go run ./tests/cmd/generate.go " + e.commit,
    "go test --tags integration ./tests"
  ]
  j.run()
})
