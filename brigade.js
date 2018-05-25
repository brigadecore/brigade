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
      let parts = gh.ref.split("/", 3)
      let tag = parts[2]
      return acrBuild(e, project, tag).then(() => {
        releaseBrig(e, project, tag)
      })
    }
    return Promise.resolve(runRelease)
  }).catch(err => {
    return ghNotify("failure", `failed build ${ err.toString() }`, e, project).run()
  });
}

function releaseBrig(e, p, tag) {
  if (!p.secrets.ghToken) {
    throw new Error("Project must have 'secrets.ghToken' set")
  }

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
    `cd /src`,
    `git checkout ${tag}`,
    // Need to move the source into GOPATH so vendor/ works as desired.
    `mkdir -p ${localPath}`,
    `cp -a /src/* ${localPath}`,
    `cp -a /src/.git ${localPath}`,
    `cd ${localPath}`,
    "dep ensure",
    "make build-release",
    `github-release release -t ${tag} -n "${parts[1]} ${tag}" || echo "release ${tag} exists"`
  ];

  // Upload for each target that we support
  for (const f of ["linux-amd64", "windows-amd64", "darwin-amd64"]) {
    var name = binName + "-"+f
    if (f == "windows-amd64") {
      name += ".exe"
    }
    cx.tasks.push(`github-release upload -f ./bin/${name} -n ${name} -t ${tag}`)  
  }
  console.log(cx.tasks);
  console.log(`releases at https://github.com/${p.repo.name}/releases/tag/${tag}`);
  return cx.run();
}

function ghNotify(state, msg, e, project) {
  const gh = new Job(`notify-${ state }`, "technosophos/github-notify:latest")
  gh.env = {
    GH_REPO: project.repo.name,
    GH_STATE: state,
    GH_DESCRIPTION: msg,
    GH_CONTEXT: "brigade",
    GH_TOKEN: project.secrets.ghToken,
    GH_COMMIT: e.revision.commit
  }
  return gh
}

function acrBuild(e, project, tag) {
  const gopath = "/go"
  const localPath = gopath + "/src/github.com/" + project.repo.name;
  const images = [
    "brig",
    "brigade-api",
    "brigade-controller",
    "brigade-cr-gateway",
    "brigade-vacuum",
    "brigade-worker", // brigade-worker does not have a rootfs. Could probably minify src into one and save space
    "git-sidecar",
    "brigade-github-gateway"
  ]

  // We build in a separate pod b/c AKS's Docker is too old to do multi-stage builds.
  const goBuild = new Job("brigade-build", goImg);
  goBuild.storage.enabled = true;
  goBuild.env = {
    "DEST_PATH": localPath,
    "GOPATH": gopath
  };
  goBuild.tasks = [
    `cd /src && git checkout ${tag}`,
    "go get github.com/golang/dep/cmd/dep",
    `mkdir -p ${localPath}/bin`,
    `mv /src/* ${localPath}`,
    `cd ${localPath}`,
    "dep ensure",
    "make build-docker-bins"
  ];

  for (let i of images) {
    goBuild.tasks.push(
      // Copy the Docker rootfs of each binary into shared storage. This is
      // a little tricky because worker is non-Go, so later we will have
      // to copy them back.
      `mkdir -p /mnt/brigade/share/${i}/rootfs`,
      // If there's no rootfs, we're done. Otherwise, copy it.
      `[ ! -d ${i}/rootfs ] || cp -a ./${i}/rootfs/* /mnt/brigade/share/${i}/rootfs/`,
    );
  }
  goBuild.tasks.push("ls -lah /mnt/brigade/share");

  let registry = "brigade"
  var builder = new Job("az-build", "microsoft/azure-cli:latest")
  builder.storage.enabled = true;
  builder.tasks = [
    // Create a service principal and assign it proper perms on the ACR.
    `az login --service-principal -u ${project.secrets.acrName} -p '${project.secrets.acrToken}' --tenant ${project.secrets.acrTenant}`,
    `cd /src`,
    `mkdir -p ./bin`
  ]

  // For each image we want to build, build it, then tag it latest, then post it to registry.
  for (let i of images) {
    let imgName = i+":"+tag;
    let latest = i+":latest";
    builder.tasks.push(
      `cd ${i}`,
      `echo '========> Building ${i}'`,
      `cp -av /mnt/brigade/share/${i}/rootfs ./rootfs`,
      `az acr build -r ${registry} -t ${imgName} -t ${latest} .`,
      `echo '<======== Finished ${i}'`,
      `cd ..`
    );
  }
  return Group.runEach([goBuild, builder]);
}

events.on("push", build)
events.on("pull_request", build)

events.on("release_brig", (e, p) => {
  /*
   * Expects JSON of the form {'tag': 'v1.2.3'}
   */
  payload = JSON.parse(e.payload)
  if (!payload.tag) {
    throw error("No tag specified")
  }

  releaseBrig(e, p, payload.tag)
})

events.on("release_images", (e, p) => {
  /*
   * Expects JSON of the form {'tag': 'v1.2.3'}
   */
  payload = JSON.parse(e.payload)
  if (!payload.tag) {
    throw error("No tag specified")
  }
  acrBuild(e, p, payload.tag)
})

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
