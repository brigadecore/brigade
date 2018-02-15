// ============================================================================
// NOTE: This is the actual brigade.js file for testing the Brigade project.
// Be careful when editing!
// ============================================================================
const { events, Job, Group} = require("brigadier")

const goImg = "golang:1.9"

function build(e, project) {
  // This is a Go project, so we want to set it up for Go.
  var gopath = "/go"

  // To set GOPATH correctly, we have to override the default
  // path that Brigade sets.

  var localPath = gopath + "/src/github.com/" + project.repo.name;

  // Create a new job to run Go tests
  var goBuild = new Job("brigade-test", goImg);

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

  start = ghNotify("pending", "Build started", e, project)

  // Run tests in parallel. Then if it's a release, push binaries.
  // Then send GitHub a notification on the status.
  Group.runAll([start, jsTest, goBuild])
  .then(() => {
      return ghNotify("success", "Passed", e, project).run()
   }).then( () => {
    const gh = JSON.parse(e.payload)
    var runRelease = false
    if (e.event == "push" && gh.ref.startsWith("refs/tags/")) {
      // Run the release in the background.
      runRelease = true
      release(e, p)
    }
    return Promise.resolve(runRelease)
  }).catch(e => {
    return ghNotify("failure", `failed build ${ e.toString() }`, e, project).run()
  });
}

function release(e, p) {
  if (!p.secrets.ghToken) {
    throw new Error("Project must have 'secrets.ghToken' set")
  }

  const tag = e.commit
  const binName = "brig"
  const gopath = "/go"
  const localPath = gopath + "/src/github.com/" + p.repo.name;

  // Cross-compile binaries for a given release and upload them to GitHub.
  var cx = new Job("cross-compile", goImg)
  cx.storage.enabled = true

  parts = p.repo.name.split("/", 2)

  cx.env = {
    GITHUB_USER: parts[0],
    GITHUB_REPO: parts[1],
    GITHUB_TOKEN: p.secrets.ghToken,
    GOPATH: gopath
  }

  cx.tasks = [
    "go get github.com/golang/dep/cmd/dep",
    "go get github.com/aktau/github-release",
    // Need to move the source into GOPATH so vendor/ works as desired.
    "mkdir -p " + localPath,
    "cp -a /src/* " + localPath,
    "cp -a /src/.git " + localPath,
    "cd " + localPath,
    "ls -lh",
    "dep ensure",
    "make build-release",
    `github-release release -t ${tag} -n "${parts[1]} ${tag}"`
  ]

  // Upload for each target that we support
  for (const f of ["linux-amd64", "windows-amd64", "darwin-amd64"]) {
    const name = binName + "-"+f
    cx.tasks.push(`github-release upload -f ./bin/${name} -n ${name} -t ${tag}`)
  }

  console.log(cx.tasks)
  cx.run().then( res => {
    console.log(`releases at https://github.com/${p.repo.name}/releases/tag/${tag}`);
  })
}

function ghNotify(state, msg, e, project) {
  const gh = new Job(`notify-${ state }`, "technosophos/github-notify:latest")
  gh.env = {
    GH_REPO: project.repo.name,
    GH_STATE: state,
    GH_DESCRIPTION: msg,
    GH_CONTEXT: "brigade",
    GH_TOKEN: project.secrets.ghToken,
    GH_COMMIT: e.commit
  }
  return gh
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

// The "release" event will attempt to create a release for the given tag, and then attach a binary.
events.on("release", release)
