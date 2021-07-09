import { events, Event, Job, ConcurrentGroup, SerialGroup } from "@brigadecore/brigadier"

const releaseTagRegex = /^refs\/tags\/(v[0-9]+(?:\.[0-9]+)*(?:\-.+)?)$/

const goImg = "brigadecore/go-tools:v0.1.0"
const jsImg = "node:12.3.1-stretch"
const kanikoImg = "brigadecore/kaniko:v0.2.0"
const helmImg = "brigadecore/helm-tools:v0.1.0"
const localPath = "/workspaces/brigade"

// MakeTargetJob is just a job wrapper around a make target.
class MakeTargetJob extends Job {
  constructor(target: string, img: string, event: Event, env?: {[key: string]: string}) {
    super(target, img, event)
    this.primaryContainer.sourceMountPath = localPath
    this.primaryContainer.workingDirectory = localPath
    this.primaryContainer.environment = env || {}
    this.primaryContainer.environment["SKIP_DOCKER"] = "true"
    if (event.worker?.git?.ref) {
      const matchStr = event.worker.git.ref.match(releaseTagRegex)
      if (matchStr) {
        this.primaryContainer.environment["VERSION"] = Array.from(matchStr)[1] as string
      }
    }
    this.primaryContainer.command = [ "make" ]
    this.primaryContainer.arguments = [ target ]
  }
}

// PushImageJob is a specialized job type for publishing Docker images.
class PushImageJob extends MakeTargetJob {
  constructor(target: string, event: Event) {
    super(target, kanikoImg, event, {
      "DOCKER_ORG": event.project.secrets.dockerhubOrg,
      "DOCKER_USERNAME": event.project.secrets.dockerhubUsername,
      "DOCKER_PASSWORD": event.project.secrets.dockerhubPassword
    })
  }
}

// A map of all jobs. When a check_run:rerequested event wants to re-run a
// single job, this allows us to easily find that job by name.
const jobs: {[key: string]: (event: Event) => Job } = {}

// Basic tests:

const testUnitGoJobName = "test-unit-go"
const testUnitGoJob = (event: Event) => {
  return new MakeTargetJob(testUnitGoJobName, goImg, event)
}
jobs[testUnitGoJobName] = testUnitGoJob

const lintGoJobName = "lint-go"
const lintGoJob = (event: Event) => {
  return new MakeTargetJob(lintGoJobName, goImg, event)
}
jobs[lintGoJobName] = lintGoJob

const testUnitJSJobName = "test-unit-js"
const testUnitJSJob = (event: Event) => {
  return new MakeTargetJob(testUnitJSJobName, jsImg, event)
}
jobs[testUnitJSJobName] = testUnitJSJob

const lintJSJobName = "lint-js"
const lintJSJob = (event: Event) => {
  return new MakeTargetJob(lintJSJobName, jsImg, event)
}
jobs[lintJSJobName] = lintJSJob

const yarnAuditJobName = "yarn-audit"
const yarnAuditJob = (event: Event) => {
  return new MakeTargetJob(yarnAuditJobName, jsImg, event)
}
jobs[yarnAuditJobName] = yarnAuditJob

const lintChartJobName = "lint-chart"
const lintChartJob = (event: Event) => {
  return new MakeTargetJob(lintChartJobName, helmImg, event)
}
jobs[lintChartJobName] = lintChartJob

const validateSchemasJobName = "validate-schemas"
const validateSchemasJob = (event: Event) => {
  return new MakeTargetJob(validateSchemasJobName, jsImg, event);
}
jobs[validateSchemasJobName] = validateSchemasJob

const validateExamplesJobName = "validate-examples";
const validateExamplesJob = (event: Event) => {
  return new MakeTargetJob(validateExamplesJobName, jsImg, event)
}
jobs[validateExamplesJobName] = validateExamplesJob;

// Build / publish stuff:

const buildAPIServerJobName = "build-apiserver"
const buildAPIServerJob = (event: Event) => {
  return new MakeTargetJob(buildAPIServerJobName, kanikoImg, event)
}
jobs[buildAPIServerJobName] = buildAPIServerJob

const pushAPIServerJobName = "push-apiserver"
const pushAPIServerJob = (event: Event) => {
  return new PushImageJob(pushAPIServerJobName, event)
}
jobs[pushAPIServerJobName] = pushAPIServerJob

const buildGitInitializerJobName = "build-git-initializer"
const buildGitInitializerJob = (event: Event) => {
  return new MakeTargetJob(buildGitInitializerJobName, kanikoImg, event)
}
jobs[buildGitInitializerJobName] = buildGitInitializerJob

const pushGitInitializerJobName = "push-git-initializer"
const pushGitInitializerJob = (event: Event) => {
  return new PushImageJob(pushGitInitializerJobName, event)
}
jobs[pushGitInitializerJobName] = pushGitInitializerJob

const buildLoggerLinuxJobName = "build-logger-linux"
const buildLoggerLinuxJob = (event: Event) => {
  return new MakeTargetJob(buildLoggerLinuxJobName, kanikoImg, event)
}
jobs[buildLoggerLinuxJobName] = buildLoggerLinuxJob

const pushLoggerLinuxJobName = "push-logger-linux"
const pushLoggerLinuxJob = (event: Event) => {
  return new PushImageJob(pushLoggerLinuxJobName, event)
}
jobs[pushLoggerLinuxJobName] = pushLoggerLinuxJob

const buildObserverJobName = "build-observer"
const buildObserverJob = (event: Event) => {
  return new MakeTargetJob(buildObserverJobName, kanikoImg, event)
}
jobs[buildObserverJobName] = buildObserverJob

const pushObserverJobName = "push-observer"
const pushObserverJob = (event: Event) => {
  return new PushImageJob(pushObserverJobName, event)
}
jobs[pushObserverJobName] = pushObserverJob

const buildSchedulerJobName = "build-scheduler"
const buildSchedulerJob = (event: Event) => {
  return new MakeTargetJob(buildSchedulerJobName, kanikoImg, event)
}
jobs[buildSchedulerJobName] = buildSchedulerJob

const pushSchedulerJobName = "push-scheduler"
const pushSchedulerJob = (event: Event) => {
  return new PushImageJob(pushSchedulerJobName, event)
}
jobs[pushSchedulerJobName] = pushSchedulerJob

const buildWorkerJobName = "build-worker"
const buildWorkerJob = (event: Event) => {
  return new MakeTargetJob(buildWorkerJobName, kanikoImg, event)
}
jobs[buildWorkerJobName] = buildWorkerJob

const pushWorkerJobName = "push-worker"
const pushWorkerJob = (event: Event) => {
  return new PushImageJob(pushWorkerJobName, event)
}
jobs[pushWorkerJobName] = pushWorkerJob

const buildBrigadierJobName = "build-brigadier"
const buildBrigadierJob = (event: Event) => {
  return new MakeTargetJob(buildBrigadierJobName, jsImg, event)
}
jobs[buildBrigadierJobName] = buildBrigadierJob

const publishBrigadierJobName = "publish-brigadier"
const publishBrigadierJob = (event: Event) => {
  return new MakeTargetJob(publishBrigadierJobName, jsImg, event, {
    "NPM_TOKEN": event.project.secrets.npmToken
  })
}
jobs[publishBrigadierJobName] = publishBrigadierJob

const buildCLIJobName = "build-cli"
const buildCLIJob = (event: Event) => {
  return new MakeTargetJob(buildCLIJobName, goImg, event)
}
jobs[buildCLIJobName] = buildCLIJob

const publishCLIJobName = "publish-cli"
const publishCLIJob = (event: Event) => {
  return new MakeTargetJob(publishCLIJobName, goImg, event, {
    "GITHUB_ORG": event.project.secrets.githubOrg,
    "GITHUB_REPO": event.project.secrets.githubRepo,
    "GITHUB_TOKEN": event.project.secrets.githubToken
  })
}
jobs[publishCLIJobName] = publishCLIJob

const publishChartJobName = "publish-chart"
const publishChartJob = (event: Event) => {
  return new MakeTargetJob(publishChartJobName, helmImg, event, {
    "HELM_REGISTRY": event.project.secrets.helmRegistry || "ghcr.io",
    "HELM_ORG": event.project.secrets.helmOrg,
    "HELM_USERNAME": event.project.secrets.helmUsername,
    "HELM_PASSWORD": event.project.secrets.helmPassword
  })
}
jobs[publishChartJobName] = publishChartJob

// Run the entire suite of tests WITHOUT publishing anything initially. If
// EVERYTHING passes AND this was a push (merge, presumably) to the v2 branch,
// then run jobs to publish "edge" images.
async function runSuite(event: Event): Promise<void> {
  await new SerialGroup(
    new ConcurrentGroup( // Basic tests
      testUnitGoJob(event),
      lintGoJob(event),
      testUnitJSJob(event),
      lintJSJob(event),
      yarnAuditJob(event),
      lintChartJob(event),
      validateSchemasJob(event),
      validateExamplesJob(event)
    ),
    new ConcurrentGroup( // Build everything
      buildAPIServerJob(event),
      buildGitInitializerJob(event),
      buildLoggerLinuxJob(event),
      buildObserverJob(event),
      buildSchedulerJob(event),
      buildWorkerJob(event),
      buildBrigadierJob(event),  
      buildCLIJob(event)
    )
  ).run()
  if (event.worker?.git?.ref == "v2") {
    // Push "edge" images.
    //
    // npm packages MUST be semantically versioned, so we DON'T publish an
    // edge brigadier package.
    //
    // To keep our github released page tidy, we're also not publishing "edge"
    // CLI binaries.
    await new ConcurrentGroup(
      pushAPIServerJob(event),
      pushGitInitializerJob(event),
      pushLoggerLinuxJob(event),
      pushObserverJob(event),
      pushSchedulerJob(event),
      pushWorkerJob(event),
    ).run()
  }
}

// Either of these events should initiate execution of the entire test suite.
events.on("brigade.sh/github", "check_suite:requested", runSuite)
events.on("brigade.sh/github", "check_suite:rerequested", runSuite)

// This event indicates a specific job is to be re-run.
events.on("brigade.sh/github", "check_run:rerequested", async event => {
  // Check run names are of the form <project name>:<job name>, so we strip
  // event.project.id.length + 1 characters off the start of the check run name
  // to find the job name.
  const jobName = JSON.parse(event.payload).check_run.name.slice(event.project.id.length + 1)
  const job = jobs[jobName]
  if (job) {
    await job(event).run()
    return
  }
  throw new Error(`No job found with name: ${jobName}`)
})

// Pushing new commits to any branch in github triggers a check suite. Such
// events are already handled above. Here we're only concerned with the case
// wherein a new TAG has been pushed-- and even then, we're only concerned with
// tags that look like a semantic version and indicate a formal release should
// be performed.
events.on("brigade.sh/github", "push", async event => {
  const matchStr = event.worker.git.ref.match(releaseTagRegex)
  if (matchStr) {
    // This is an official release with a semantically versioned tag
    await new SerialGroup(
      new ConcurrentGroup(
        pushAPIServerJob(event),
        pushGitInitializerJob(event),
        pushLoggerLinuxJob(event),
        pushObserverJob(event),
        pushSchedulerJob(event),
        pushWorkerJob(event)
      ),
      new ConcurrentGroup(
        publishBrigadierJob(event),
        publishChartJob(event),
        publishCLIJob(event)
      )
    ).run()
  } else {
    console.log(`Ref ${event.worker.git.ref} does not match release tag regex (${releaseTagRegex}); not releasing.`)
  }
})

events.process()
