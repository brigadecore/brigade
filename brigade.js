// ============================================================================
// NOTE: This is a Brigade 1.x script for now.
// ============================================================================
const { events, Job } = require("brigadier");
const { Check } = require("@brigadecore/brigade-utils");

const goImg = "krancour/go-tools:v0.4.0";
const dockerImg = "docker:stable-dind";
const localPath = "/workspaces/brigade";

// Run Go unit tests
function testUnitGo() {
  var job = new Job("test-unit-go", goImg);
  job.mountPath = localPath;
  job.env = {
    "SKIP_DOCKER": "true"
  };
  job.tasks = [
    `cd ${localPath}`,
    "make test-unit-go"
  ];
  return job;
}

// Run Go lint checks
function lintGo() {
  var job = new Job("lint-go", goImg);
  job.mountPath = localPath;
  job.env = {
    "SKIP_DOCKER": "true"
  };
  
  job.tasks = [
    `cd ${localPath}`,
    "make lint-go"
  ];
  return job;
}

// Build the API server
function buildAPIServer() {
  return buildImage("apiserver");
}

// Build the scheduler
function buildScheduler() {
  return buildImage("scheduler");
}

// Build the observer
function buildObserver() {
  return buildImage("observer");
}

// Build the Linux logging agent
function buildLoggerLinux() {
  return buildImage("logger-linux");
}

function buildImage(imageName) {
  var job = new Job(`build-${imageName}`, dockerImg);
  job.mountPath = localPath;
  job.privileged = true;
  job.tasks = [
    "apk add --update --no-cache make git",
    "dockerd-entrypoint.sh &",
    "sleep 20",
    `cd ${localPath}`,
    `make build-${imageName}`
  ];
  return job;
}

// Build the CLI
function buildCLI() {
  var job = new Job("build-cli", goImg);
  job.mountPath = localPath;
  job.env = {
    "SKIP_DOCKER": "true"
  };
  job.tasks = [
    `cd ${localPath}`,
    "make xbuild-cli"
  ];
  return job;
}

function runSuite(e, p) {
  // Important: To prevent Promise.all() from failing fast, we catch and
  // return all errors. This ensures Promise.all() always resolves. We then
  // iterate over all resolved values looking for errors. If we find one, we
  // throw it so the whole build will fail.
  //
  // Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all#Promise.all_fail-fast_behaviour
  return Promise.all([
    run(e, p, testUnitGo).catch((err) => { return err }),
    run(e, p, lintGo).catch((err) => { return err }),
    run(e, p, buildAPIServer).catch((err) => { return err }),
    run(e, p, buildScheduler).catch((err) => { return err }),
    run(e, p, buildObserver).catch((err) => { return err }),
    run(e, p, buildLoggerLinux).catch((err) => { return err }),
    run(e, p, buildCLI).catch((err) => { return err })
  ]).then((values) => {
    values.forEach((value) => {
      if (value instanceof Error) throw value;
    });
  });
}

function runCheck(e, p) {
  payload = JSON.parse(e.payload);
  name = payload.body.check_run.name;
  // Determine which check to run
  switch (name) {
    case "test-unit-go":
      return run(e, p, testUnitGo);
    case "lint-go":
      return run(e, p, lintGo);
    case "build-apiserver":
      return run(e, p, buildAPIServer);
    case "build-scheduler":
      return run(e, p, buildScheduler);
    case "build-observer":
      return run(e, p, buildObserver);
    case "build-logger-linux":
      return run(e, p, buildLoggerLinux);
    case "build-cli":
      return run(e, p, buildCLI);
    default:
      throw new Error(`No check found with name: ${name}`);
  }
}

// run is a Check Run that is run as part of a Checks Suite
function run(e, p, jobFunc) {
  console.log("Check requested");
  var check = new Check(e, p, jobFunc(), `https://brigadecore.github.io/kashti/builds/${e.buildID}`);
  return check.run();
}

events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);
events.on("check_run:rerequested", runCheck);
events.on("issue_comment:created", (e, p) => Check.handleIssueComment(e, p, runSuite));
events.on("issue_comment:edited", (e, p) => Check.handleIssueComment(e, p, runSuite));
