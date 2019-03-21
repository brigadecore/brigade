// ============================================================================
// NOTE: This is the actual brigade.js file for testing the Brigade project.
// Be careful when editing!
// ============================================================================
const { events, Job, Group } = require("brigadier");

const projectName = "brigade";
const projectOrg = "brigadecore";

// Go build defaults
const goImg = "golang:1.11";
const gopath = "/go";
const localPath = gopath + `/src/github.com/${projectOrg}/${projectName}`;
const goEnv = {
  "DEST_PATH": localPath,
  "GOPATH": gopath
};

// Used by Docker image build/publish jobs
const sharedMountPrefix = `/mnt/${projectName}/share/`;
const addMake =
  "apk upgrade 1>/dev/null && \
  apk add --update --no-cache make 1>/dev/null";

const noop = {run: () => {return Promise.resolve()}};

function goTest(e, project) {
  // Create a new job to run Go tests
  var goTest = new Job("go-test", goImg);

  goTest.env = goEnv;
  goTest.tasks = [
    // Need to move the source into GOPATH so vendor/ works as desired.
    "mkdir -p " + localPath,
    "mv /src/* " + localPath,
    "mv /src/.git " + localPath,
    "cd " + localPath,
    "make vendor",
    "make test-style",
    "make test-unit"
  ];

  return goTest;
}

function jsTest(e, project) {
  // Run the javascript-based brigade worker tests
  var jsTest = new Job("js-test", "node:8");

  jsTest.tasks = [
    "cd /src",
    "make bootstrap-js",
    "make test-js"
  ];

  return jsTest;
}

// Here we can add additional Check Runs, which will run in parallel and
// report their results independently to GitHub
function runSuite(e, p) {
  // Note: as provided language string is used in job naming, it must consist
  // of lowercase letters and hyphens only (per Brigade/K8s restrictions)
  checkRun(e, p, goTest, "go").catch(e  => {console.error(e.toString())});
  checkRun(e, p, jsTest, "javascript").catch(e  => {console.error(e.toString())});
}

// checkRun is a GitHub Check Run that is ran as part of a Checks Suite,
// running the provided runFunc with language-based messaging
function checkRun(e, p, runFunc, language) {
  console.log("Check requested");

  // Create Notification object (which is just a Job to update GH using the Checks API)
  var note = new Notification(`test-${language}`, e, p);
  note.conclusion = "";
  note.title = `Run ${language} tests`;
  note.summary = `Running the ${language} build/test targets for ${e.revision.commit}`;
  note.text = "This task will ensure build, linting and tests all pass.";

  // Send notification, then run, then send pass/fail notification
  return notificationWrap(runFunc(e, p), note);
}

// A GitHub Check Suite notification
class Notification {
  constructor(name, e, p) {
    this.proj = p;
    this.payload = e.payload;
    this.name = name;
    this.externalID = e.buildID;
    this.detailsURL = `https://brigadecore.github.io/kashti/builds/${ e.buildID }`;
    this.title = "running check";
    this.text = "";
    this.summary = "";

    // count allows us to send the notification multiple times, with a distinct pod name
    // each time.
    this.count = 0;

    // One of: "success", "failure", "neutral", "cancelled", or "timed_out".
    this.conclusion = "neutral";
  }

  // Send a new notification, and return a Promise<result>.
  run() {
    this.count++;
    var j = new Job(`${ this.name }-${ this.count }`, "deis/brigade-github-check-run:latest");
    j.imageForcePull = true;
    j.env = {
      CHECK_CONCLUSION: this.conclusion,
      CHECK_NAME: this.name,
      CHECK_TITLE: this.title,
      CHECK_PAYLOAD: this.payload,
      CHECK_SUMMARY: this.summary,
      CHECK_TEXT: this.text,
      CHECK_DETAILS_URL: this.detailsURL,
      CHECK_EXTERNAL_ID: this.externalID
    };
    return j.run();
  }
}

// Helper to wrap a job execution between two notifications.
async function notificationWrap(job, note) {
  await note.run();
  try {
    let res = await job.run();
    const logs = await job.logs();

    note.conclusion = "success";
    note.summary = `Task "${ job.name }" passed`;
    note.text = note.text = "```" + res.toString() + "```\nComplete";
    return await note.run();
  } catch (e) {
    const logs = await job.logs();
    note.conclusion = "failure";
    note.summary = `Task "${ job.name }" failed for ${ e.buildID }`;
    note.text = "```" + logs + "```\nFailed with error: " + e.toString();
    try {
      return await note.run();
    } catch (e2) {
      console.error("failed to send notification: " + e2.toString());
      console.error("original error: " + e.toString());
      return e2;
    }
  }
}

function release(p, tag) {
  if (!p.secrets.ghToken) {
    throw new Error("Project must have 'secrets.ghToken' set");
  }

  // Cross-compile binaries for a given release and upload them to GitHub.
  var cx = new Job("release", goImg);

  cx.storage.enabled = true;
  parts = p.repo.name.split("/", 2);
  cx.env = {
    GITHUB_USER: parts[0],
    GITHUB_REPO: parts[1],
    GITHUB_TOKEN: p.secrets.ghToken,
    GOPATH: gopath
  };

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
    `last_tag=$(git describe --tags ${tag}^ --abbrev=0 --always)`,
    `github-release release \
      -t ${tag} \
      -n "${parts[1]} ${tag}" \
      -d "$(git log --no-merges --pretty=format:'- %s %H (%aN)' HEAD ^$last_tag)" \
      || echo "release ${tag} exists"`,
    "for bin in ./bin/*; do github-release upload -f ${bin} -n $(basename ${bin}) -t " + tag + "; done"
  ];

  console.log(cx.tasks);
  console.log(`releases at https://github.com/${p.repo.name}/releases/tag/${tag}`);
  return cx;
}

// Separate docker build stage as there may be multiple consumers/publishers,
// For example, publishing to Dockerhub below
function goDockerBuild(e, p) {
  // We build in a separate pod b/c AKS's Docker is too old to do multi-stage builds.
  const builder = new Job(`${projectName}-docker-build`, goImg);

  builder.storage.enabled = true;
  builder.env = goEnv;
  builder.tasks = [
    `cd /src`,
    `mkdir -p ${localPath}/bin`,
    `cp -a /src/* ${localPath}`,
    `cp -a /src/.git ${localPath}`,
    `cd ${localPath}`,
    "make vendor",
    "make build-docker-bins",
    // Copy the Docker rootfs of each binary into shared storage. This is
    // a little tricky because worker is non-Go, so later we will have
    // to copy them back.
    "for i in $(make echo-images); do \
        mkdir -p " + sharedMountPrefix + "${i}/rootfs; \
        [ ! -d ${i}/rootfs ] || cp -a ./${i}/rootfs/* " + sharedMountPrefix + "${i}/rootfs/; \
      done",
    `ls -lah ${sharedMountPrefix}`
  ];

  return builder;
}

function dockerhubPublish(project, tag) {
  const publisher = new Job("dockerhub-publish", "docker");
  let dockerRegistry = project.secrets.dockerhubRegistry || "docker.io";
  let dockerOrg = project.secrets.dockerhubOrg || "deis";

  publisher.docker.enabled = true;
  publisher.storage.enabled = true;
  publisher.tasks = [
    addMake,
    `docker login ${dockerRegistry} -u ${project.secrets.dockerhubUsername} -p ${project.secrets.dockerhubPassword}`,
    "cd /src",
    "for i in $(make echo-images); do \
       cp -av " + sharedMountPrefix + "${i}/rootfs ./${i}; \
       DOCKER_REGISTRY=" + dockerOrg + " VERSION=" + tag + " make ${i}-image ${i}-push; \
     done",
    `docker logout ${dockerRegistry}`
  ];

  return publisher;
}

function slackNotify(title, msg, project) {
  if (project.secrets.SLACK_WEBHOOK) {
    var slack = new Job(`${projectName}-slack-notify`, "technosophos/slack-notify:latest");

    slack.env = {
      SLACK_WEBHOOK: project.secrets.SLACK_WEBHOOK,
      SLACK_USERNAME: "brigade-ci",
      SLACK_TITLE: title,
      SLACK_MESSAGE: msg,
      SLACK_COLOR: "#00ff00"
    };
    slack.tasks = ["/slack-notify"];

    return slack;
  } else {
    console.log(`Slack Notification for '${title}' not sent; no SLACK_WEBHOOK secret found.`);
    return noop;
  }
}

events.on("exec", (e, p) => {
  return Group.runAll([
    goTest(e, p),
    jsTest(e, p)
  ]);
});

// Although a GH App will trigger 'check_suite:requested' on a push to master event,
// it will not for a tag push, hence the need for this handler
events.on("push", (e, p) => {
  let doPublish = false;
  let doRelease = false;
  let tag = "";
  let jobs = [];

  if (e.revision.ref.includes("refs/heads/master")) {
    doPublish = true;
    tag = "latest";
  } else if (e.revision.ref.startsWith("refs/tags/")) {
    doPublish = true;
    doRelease = true;
    let parts = e.revision.ref.split("/", 3);
    tag = parts[2];
  }

  if (doPublish) {
    jobs.push(
      goDockerBuild(e, p),
      dockerhubPublish(p, tag)
    );
  }

  if (doRelease) {
    jobs.push(
      release(p, tag),
      slackNotify(
        "Brigade Release", 
        `${tag} release now on GitHub! <https://github.com/${p.repo.name}/releases/tag/${tag}>`, 
        p
      )
    );
  }

  if (jobs.length) {
    Group.runEach(jobs);
  }
});

events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);
events.on("check_run:rerequested", runSuite);

events.on("release", (e, p) => {
  /*
   * Expects JSON of the form {'tag': 'v1.2.3'}
   */
  payload = JSON.parse(e.payload);
  if (!payload.tag) {
    throw error("No tag specified");
  }

  release(p, payload.tag).run();
});

events.on("publish", (e, p) => {
  /*
   * Expects JSON of the form {'tag': 'v1.2.3'}
   */
  payload = JSON.parse(e.payload);
  if (!payload.tag) {
    throw error("No tag specified");
  }

  Group.runEach([
    goDockerBuild(e, p),
    dockerhubPublish(p, payload.tag)
  ]);
});
