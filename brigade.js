// ============================================================================
// NOTE: This is a Brigade 1.x script for now.
// ============================================================================

const { events } = require("brigadier");
const { Check, KindJob } = require("@brigadecore/brigade-utils");

const localPath = "/workspaces/brigade";

// A map of all jobs. When a check_run:rerequested event wants to re-run a
// single job, this allows us to easily find that job by name.
const jobs = {};

const testIntegrationJobName = "test-integration";
const testIntegrationJob = (e, p) => {
  let kind = new KindJob(testIntegrationJobName);
  kind.mountPath = localPath;
  kind.env = {
    "BRIGADE_CI_PRIVATE_REPO_SSH_KEY": p.secrets.privateRepoSSHKey,
    // IMAGE_PULL_POLICY is set to 'IfNotPresent' for deploy; if set to
    // 'Always', the images loaded manually into kind will be ignored and the
    // pods will attempt to pull from remote registries.
    "IMAGE_PULL_POLICY": "IfNotPresent",
    // Use a hard-coded apiserver password for deployment and subsequent login
    "APISERVER_ROOT_PASSWORD": "F00Bar!!!"
  };
  kind.tasks.push(
    // Install git and golang deps
    "apk add --update --no-cache git gcc musl-dev",
    // Install helm
    `curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 \
      && chmod 700 get_helm.sh \
      && ./get_helm.sh`,
    `cd ${localPath}`,
    // It would be more efficient to mount built images from associated
    // build jobs; however, they currently use kaniko, which doesn't preserve
    // images in the local Docker cache.
    "make hack-build-images hack-load-images",
    "make hack-deploy test-integration",
  );
  return kind;
}
jobs[testIntegrationJobName] = testIntegrationJob;

// Run the entire suite of tests, builds, etc. concurrently WITHOUT publishing
// anything initially. If EVERYTHING passes AND this was a push (merge,
// presumably) to the v2 branch, then run jobs to publish "edge" images.
function runSuite(e, p) {
  // Important: To prevent Promise.all() from failing fast, we catch and
  // return all errors. This ensures Promise.all() always resolves. We then
  // iterate over all resolved values looking for errors. If we find one, we
  // throw it so the whole build will fail.
  //
  // Ref: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Promise/all#Promise.all_fail-fast_behaviour
  return Promise.all([
    run(e, p, testIntegrationJob(e, p)).catch((err) => { return err })
  ]).then((values) => {
    values.forEach((value) => {
      if (value instanceof Error) throw value;
    });
  });
}

// run the specified job, sandwiched between two other jobs to report status
// via the GitHub checks API.
function run(e, p, job) {
  console.log("Check requested");
  var check = new Check(e, p, job, `https://brigadecore.github.io/kashti/builds/${e.buildID}`);
  return check.run();
}

// Either of these events should initiate execution of the entire test suite.
events.on("check_suite:requested", runSuite);
events.on("check_suite:rerequested", runSuite);

// These events MAY indicate that a maintainer has expressed, via a comment,
// that the entire test suite should be run.
events.on("issue_comment:created", (e, p) => Check.handleIssueComment(e, p, runSuite));
events.on("issue_comment:edited", (e, p) => Check.handleIssueComment(e, p, runSuite));

// This event indicates a specific job is to be re-run.
events.on("check_run:rerequested", (e, p) => {
  const jobName = JSON.parse(e.payload).body.check_run.name;
  const job = jobs[jobName];
  if (job) {
    return run(e, p, job(e, p));
  }
  throw new Error(`No job found with name: ${jobName}`);
});
