// ============================================================================
// NOTE: This is the actual brigade.js file for testing the Brigade project.
// Be careful when editing!
// ============================================================================
const { events, Job, Group} = require("brigadier")

const goImg = "golang:1.11"

// TODO: Possible/preferred to use canonical image list in Makefile?
// Would necessitate looping inside of Job, i.e., within one shell task
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
    // Need to move the source into GOPATH so vendor/ works as desired.
    "mkdir -p " + localPath,
    "mv /src/* " + localPath,
    "cd " + localPath,
    "make vendor",
    "make test-style",
    "make test-unit"
  ];

  // Run the brigade worker tests
  var jsTest = new Job("brigade-js-build", "node:8");
  jsTest.tasks = [
    "cd /src/brigade-worker",
    "yarn install",
    "yarn test"
  ];

  start = ghNotify("pending", `Build started as ${ e.buildID }`, e, project)

  // Run tests in parallel. Then if it's a release, push binaries.
  // Then send GitHub a notification on the status.
  Group.runAll([start, jsTest, goBuild])
  .then(() => {
      return ghNotify("success", `Build ${ e.buildID } passed`, e, project).run()
   }).then( () => {
    var runRelease = false
    if (e.type == "push") {
      // Push to master: run "latest" release, don't release brig
      if (e.revision.ref.includes("refs/heads/master")) {
        runRelease = true
        return goDockerBuild(project, e.revision.ref).run()
          .then(() => {
            Group.runAll([
              acrBuild(project, "latest"),
              dockerhubPublish(project, "latest")
            ])
          })
      }
      // Tag pushed: run tag release, release brig
      if (e.revision.ref.startsWith("refs/tags/")) {
        runRelease = true
        let parts = e.revision.ref.split("/", 3)
        let tag = parts[2]
        return goDockerBuild(project, tag).run()
          .then(() => {
            Group.runAll([
              acrBuild(project, tag),
              dockerhubPublish(project, tag)
            ])
          })
          .then(() => {
            releaseBrig(project, tag)
          })
      }
    }
    return Promise.resolve(runRelease)
  }).catch(err => {
    return ghNotify("failure", `failed build ${ e.buildID }`, e, project).run()
  });
}

function releaseBrig(p, tag) {
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
    "go get github.com/aktau/github-release",
    `cd /src`,
    `git checkout ${tag}`,
    // Need to move the source into GOPATH so vendor/ works as desired.
    `mkdir -p ${localPath}`,
    `cp -a /src/* ${localPath}`,
    `cp -a /src/.git ${localPath}`,
    `cd ${localPath}`,
    "make vendor",
    "make build-release",
    `github-release release -t ${tag} -n "${parts[1]} ${tag}" || echo "release ${tag} exists"`
  ];

  // Upload for each target that we support
  for (const f of ["linux-amd64", "windows-amd64.exe", "darwin-amd64"]) {
    var name = binName + "-" + f;
    var outname = name;
    cx.tasks.push(`github-release upload -f ./bin/${name} -n ${outname} -t ${tag}`)  
  }
  console.log(cx.tasks);
  console.log(`releases at https://github.com/${p.repo.name}/releases/tag/${tag}`);
  return cx.run();
}

function ghNotify(state, msg, e, project) {
  const gh = new Job(`notify-${ state }`, "technosophos/github-notify:latest")
  gh.imageForcePull = true;
  gh.env = {
    GH_REPO: project.repo.name,
    GH_STATE: state,
    GH_DESCRIPTION: msg,
    GH_CONTEXT: "brigade",
    GH_TOKEN: project.secrets.ghToken,
    GH_COMMIT: e.revision.commit,
    GH_TARGET_URL: `https://azure.github.io/kashti/builds/${ e.buildID }`,
  }
  return gh
}

function goDockerBuild(project, tag) {
  // We build in a separate pod b/c AKS's Docker is too old to do multi-stage builds.
  const goBuild = new Job("brigade-docker-build", goImg);
  const gopath = "/go"
  const localPath = gopath + "/src/github.com/" + project.repo.name;

  goBuild.storage.enabled = true;
  goBuild.env = {
    "DEST_PATH": localPath,
    "GOPATH": gopath
  };
  goBuild.tasks = [
    `cd /src && git checkout ${tag}`,
    `mkdir -p ${localPath}/bin`,
    `mv /src/* ${localPath}`,
    `cd ${localPath}`,
    "make vendor",
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

  return goBuild;
}

function dockerhubPublish(project, tag) {
  const publisher = new Job("dockerhub-publish", "docker");
  let dockerRegistry = project.secrets.dockerhubRegistry || "docker.io";
  let dockerOrg = project.secrets.dockerhubOrg || "deis";

  publisher.docker.enabled = true;
  publisher.storage.enabled = true;
  publisher.tasks = [
    "apk add --update --no-cache make",
    `docker login ${dockerRegistry} -u ${project.secrets.dockerhubUsername} -p ${project.secrets.dockerhubPassword}`,
    "cd /src"
  ];

  for (let i of images) {
      publisher.tasks.push(
        `cp -av /mnt/brigade/share/${i}/rootfs ./${i}`,
        `DOCKER_REGISTRY=${dockerOrg} VERSION=${tag} make ${i}-image ${i}-push`
      );
  }
  publisher.tasks.push(`docker logout ${dockerRegistry}`);

  return publisher;
}

function acrBuild(project, tag) {
  const acrImagePrefix = "public/deis/"

  let registry = project.secrets.acrRegistry || "brigade"
  var builder = new Job("az-build", "microsoft/azure-cli:latest")
  builder.imageForcePull = true;
  builder.storage.enabled = true;
  builder.tasks = [
    // Create a service principal and assign it proper perms on the ACR.
    `az login --service-principal -u ${project.secrets.acrName} -p '${project.secrets.acrToken}' --tenant ${project.secrets.acrTenant}`,
    `cd /src`,
    `mkdir -p ./bin`
  ]

  // For each image, build, tag and finally push to registry
  for (let i of images) {
    let imgName = acrImagePrefix+i;
    let tagged = imgName+":"+tag;

    builder.tasks.push(
      `cd ${i}`,
      `echo '========> Building ${i}'`,
      `cp -av /mnt/brigade/share/${i}/rootfs ./`,
      `az acr build -r ${registry} -t ${tagged} .`,
      `echo '<======== Finished ${i}'`,
      `cd ..`
    );
  }
  return builder;
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

  releaseBrig(p, payload.tag)
})

events.on("release_images", (e, p) => {
  /*
   * Expects JSON of the form {'tag': 'v1.2.3'}
   */
  payload = JSON.parse(e.payload)
  if (!payload.tag) {
    throw error("No tag specified")
  }

  goDockerBuild(p, payload.tag).run()
    .then(() => {
      Group.runAll([
        acrBuild(p, payload.tag),
        dockerhubPublish(p, payload.tag)
      ])
    });
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
