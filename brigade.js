// ============================================================================
// NOTE: This is the actual brigade.js file for testing the Brigade project.
// Be careful when editing!
// ============================================================================
const { events, Job, Group } = require("brigadier");

const projectName = "brigade";
const projectOrg = "brigadecore";

// Go build defaults
const goImg = "quay.io/deis/lightweight-docker-go:v0.6.0";
const gopath = "/go";
const localPath = gopath + `/src/github.com/${projectOrg}/${projectName}`;

// JS build defaults
const jsImg = "node:12.3.1-stretch";

const releaseTagRegex = /^refs\/tags\/(v[0-9]+(?:\.[0-9]+)*(?:\-.+)?)$/;

const noopJob = {run: () => {return Promise.resolve()}};

function goTest() {
  // Create a new job to run Go tests
  var job = new Job("go-test", goImg);
  job.mountPath = localPath;
  // Set a few environment variables.
  job.env = {
    "SKIP_DOCKER": "true"
  };
  // Run Go unit tests
  job.tasks = [
    `cd ${localPath}`,
    "make verify-vendored-code lint test-unit"
  ];
  return job;
}

function jsTest() {
  // Create a new job to run JS-based Brigade worker tests
  var job = new Job("js-test", jsImg);
  // Set a few environment variables.
  job.env = {
    "SKIP_DOCKER": "true"
  };
  job.tasks = [
    "cd /src",
    "make verify-vendored-code-js test-js"
  ];
  return job;
}

function buildAndPublishImages(project, version) {
  let dockerRegistry = project.secrets.dockerhubRegistry || "docker.io";
  let dockerOrg = project.secrets.dockerhubOrg || "brigadecore";
  var job = new Job("build-and-publish-images", "docker:stable-dind");
  job.privileged = true;
  job.tasks = [
    "apk add --update --no-cache make git",
    "dockerd-entrypoint.sh &",
    "sleep 20",
    "cd /src",
    `docker login ${dockerRegistry} -u ${project.secrets.dockerhubUsername} -p ${project.secrets.dockerhubPassword}`,
    `DOCKER_REGISTRY=${dockerRegistry} DOCKER_ORG=${dockerOrg} VERSION=${version} make build-all-images push-all-images`,
    `docker logout ${dockerRegistry}`
  ];
  return job;
}

// Here we can add additional Check Runs, which will run in parallel and
// report their results independently to GitHub
function runSuite(e, p) {
  // For the master branch, we build and publish images in response to the push
  // event. We test as a precondition for doing that, so we DON'T test here
  // for the master branch.
  if (e.revision.ref != "master") {
    // Important: To prevent Promise.all() from failing fast, we catch and
    // return all errors. This ensures Promise.all() always resolves. We then
    // iterate over all resolved values looking for errors. If we find one, we
    // throw it so the whole build will fail.
    //
    // Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all#Promise.all_fail-fast_behaviour
    //
    // Note: as provided language string is used in job naming, it must consist
    // of lowercase letters and hyphens only (per Brigade/K8s restrictions)
    return Promise.all([
      checkRun(e, p, goTest, "go").catch((err) => {return err}),
      checkRun(e, p, jsTest, "javascript").catch((err) => {return err}),
    ])
    .then((values) => {
      values.forEach((value) => {
        if (value instanceof Error) throw value;
      });
    });
  }
}

// runTests is a Check Run that is run as part of a Checks Suite
function runTests(e, p, runFunc, language) {
  console.log("Check requested");

  // Create Notification object (which is just a Job to update GH using the Checks API)
  var note = new Notification(`test-${language}`, e, p);
  note.conclusion = "";
  note.title = `Run ${language} tests`;
  note.summary = `Running the ${language} build/test targets for ${e.revision.commit}`;
  note.text = "This task will ensure build, linting and tests all pass.";

  // Send notification, then run, then send pass/fail notification
  return notificationWrap(runFunc(), note);
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
    var job = new Job(`${ this.name }-${ this.count }`, "brigadecore/brigade-github-check-run:latest");
    job.imageForcePull = true;
    job.env = {
      "CHECK_CONCLUSION": this.conclusion,
      "CHECK_NAME": this.name,
      "CHECK_TITLE": this.title,
      "CHECK_PAYLOAD": this.payload,
      "CHECK_SUMMARY": this.summary,
      "CHECK_TEXT": this.text,
      "CHECK_DETAILS_URL": this.detailsURL,
      "CHECK_EXTERNAL_ID": this.externalID
    };
    return job.run();
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
      await note.run();
    } catch (e2) {
      console.error("failed to send notification: " + e2.toString());
      console.error("original error: " + e.toString());
    }
    throw e;
  }
}

function githubRelease(p, tag) {
  if (!p.secrets.ghToken) {
    throw new Error("Project must have 'secrets.ghToken' set");
  }
  // Cross-compile binaries for a given release and upload them to GitHub.
  var job = new Job("release", goImg);
  job.mountPath = localPath;
  parts = p.repo.name.split("/", 2);
  // Set a few environment variables.
  job.env = {
    "SKIP_DOCKER": "true",
    "GITHUB_USER": parts[0],
    "GITHUB_REPO": parts[1],
    "GITHUB_TOKEN": p.secrets.ghToken,
  };
  job.tasks = [
    "go get github.com/aktau/github-release",
    `cd ${localPath}`,
    "make build-brig",
    `last_tag=$(git describe --tags ${tag}^ --abbrev=0 --always)`,
    `github-release release \
      -t ${tag} \
      -n "${parts[1]} ${tag}" \
      -d "$(git log --no-merges --pretty=format:'- %s %H (%aN)' HEAD ^$last_tag)" \
      || echo "release ${tag} exists"`,
    `for bin in ./bin/*; do github-release upload -f $bin -n $(basename $bin) -t ${tag}; done`
  ];
  console.log(job.tasks);
  console.log(`releases at https://github.com/${p.repo.name}/releases/tag/${tag}`);
  return job;
}

function slackNotify(title, msg, project) {
  if (project.secrets.SLACK_WEBHOOK) {
    var job = new Job(`${projectName}-slack-notify`, "technosophos/slack-notify:latest");
    job.env = {
      "SLACK_WEBHOOK": project.secrets.SLACK_WEBHOOK,
      "SLACK_USERNAME": "brigade-ci",
      "SLACK_TITLE": title,
      "SLACK_MESSAGE": msg,
      "SLACK_COLOR": "#00ff00"
    };
    job.tasks = ["/slack-notify"];
    return job;
  }
  console.log(`Slack Notification for '${title}' not sent; no SLACK_WEBHOOK secret found.`);
  return noopJob;
}

events.on("exec", (e, p) => {
  return Group.runAll([
    goTest(),
    jsTest()
  ]);
});

events.on("push", (e, p) => {
  let matchStr = e.revision.ref.match(releaseTagRegex);
  if (matchStr) {
    // This is an official release with a semantically versioned tag
    let matchTokens = Array.from(matchStr);
    let version = matchTokens[1];
    return buildAndPublishImages(p, version).run()
    .then(() => {
      githubRelease(p, version).run();
    })
    .then(() => {
      slackNotify(
        "Brigade Release", 
        `${version} release now on GitHub! <https://github.com/${p.repo.name}/releases/tag/${version}>`, 
        p
      ).run();
    });
  }
  if (e.revision.ref == "refs/heads/master") {
    // This runs tests then builds and publishes "edge" images
    return Group.runAll([
      goTest(),
      jsTest()
    ])
    .then(() => {
      buildAndPublishImages(p, "").run();
    });
  }
})

events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);
events.on("check_run:rerequested", runSuite);
