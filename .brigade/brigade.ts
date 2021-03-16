import { events, Job, ConcurrentGroup } from "@brigadecore/brigadier"

// const releaseTagRegex = /^refs\/tags\/(v[0-9]+(?:\.[0-9]+)*(?:\-.+)?)$/

const goImg = "brigadecore/go-tools:v0.1.0"
const jsImg = "node:12.3.1-stretch"
const kanikoImg = "brigadecore/kaniko:v0.2.0"
const helmImg = "brigadecore/helm-tools:v0.1.0"
const localPath = "/workspaces/brigade"

// MakeTargetJob is just a job wrapper around a make target.
class MakeTargetJob extends Job {
  constructor(target: string, img: string, e: any, env?: {[key: string]: string}) {
    super(target, img, e)
    this.primaryContainer.sourceMountPath = localPath
    this.primaryContainer.workingDirectory = localPath
    this.primaryContainer.environment = env || {}
    this.primaryContainer.environment["SKIP_DOCKER"] = "true"
    // const matchStr = e.revision.ref.match(releaseTagRegex)
    // if (matchStr) {
    //   this.primaryContainer.environment["VERSION"] = Array.from(matchStr)[1] as string
    // }
    this.primaryContainer.command = [ "make" ]
    this.primaryContainer.arguments = [ target ]
  }
}

// PushImageJob is a specialized job type for publishing Docker images.
class PushImageJob extends MakeTargetJob {
  constructor(target: string, e: any) {
    super(target, kanikoImg, e, {
      "DOCKER_ORG": e.project.secrets.dockerhubOrg,
      "DOCKER_USERNAME": e.project.secrets.dockerhubUsername,
      "DOCKER_PASSWORD": e.project.secrets.dockerhubPassword
    })
  }
}

// A map of all jobs. When a check_run:rerequested event wants to re-run a
// single job, this allows us to easily find that job by name.
const jobs: {[key: string]: (event: any) => Job } = {}

// Basic tests:

const testUnitGoJobName = "test-unit-go"
const testUnitGoJob = (event: any) => {
  return new MakeTargetJob(testUnitGoJobName, goImg, event)
}
jobs[testUnitGoJobName] = testUnitGoJob

const lintGoJobName = "lint-go"
const lintGoJob = (event: any) => {
  return new MakeTargetJob(lintGoJobName, goImg, event)
}
jobs[lintGoJobName] = lintGoJob

const testUnitJSJobName = "test-unit-js"
const testUnitJSJob = (event: any) => {
  return new MakeTargetJob(testUnitJSJobName, jsImg, event)
}
jobs[testUnitJSJobName] = testUnitJSJob

const lintJSJobName = "lint-js"
const lintJSJob = (event: any) => {
  return new MakeTargetJob(lintJSJobName, jsImg, event)
}
jobs[lintJSJobName] = lintJSJob

// Brigadier:

const buildBrigadierJobName = "build-brigadier"
const buildBrigadierJob = (event: any) => {
  return new MakeTargetJob(buildBrigadierJobName, jsImg, event)
}
jobs[buildBrigadierJobName] = buildBrigadierJob

const publishBrigadierJobName = "publish-brigadier"
const publishBrigadierJob = (event: any) => {
  return new MakeTargetJob(publishBrigadierJobName, jsImg, event, {
    "NPM_TOKEN": event.project.secrets.npmToken
  })
}
jobs[publishBrigadierJobName] = publishBrigadierJob

// Docker images:

const buildAPIServerJobName = "build-apiserver"
const buildAPIServerJob = (event: any) => {
  return new MakeTargetJob(buildAPIServerJobName, kanikoImg, event)
}
jobs[buildAPIServerJobName] = buildAPIServerJob

const pushAPIServerJobName = "push-apiserver"
const pushAPIServerJob = (event: any) => {
  return new PushImageJob(pushAPIServerJobName, event)
}
jobs[pushAPIServerJobName] = pushAPIServerJob

const buildGitInitializerJobName = "build-git-initializer"
const buildGitInitializerJob = (event: any) => {
  return new MakeTargetJob(buildGitInitializerJobName, kanikoImg, event)
}
jobs[buildGitInitializerJobName] = buildGitInitializerJob

const pushGitInitializerJobName = "push-git-initializer"
const pushGitInitializerJob = (event: any) => {
  return new PushImageJob(pushGitInitializerJobName, event)
}
jobs[pushGitInitializerJobName] = pushGitInitializerJob

const buildLoggerLinuxJobName = "build-logger-linux"
const buildLoggerLinuxJob = (event: any) => {
  return new MakeTargetJob(buildLoggerLinuxJobName, kanikoImg, event)
}
jobs[buildLoggerLinuxJobName] = buildLoggerLinuxJob

const pushLoggerLinuxJobName = "push-logger-linux"
const pushLoggerLinuxJob = (event: any) => {
  return new PushImageJob(pushLoggerLinuxJobName, event)
}
jobs[pushLoggerLinuxJobName] = pushLoggerLinuxJob

const buildObserverJobName = "build-observer"
const buildObserverJob = (event: any) => {
  return new MakeTargetJob(buildObserverJobName, kanikoImg, event)
}
jobs[buildObserverJobName] = buildObserverJob

const pushObserverJobName = "push-observer"
const pushObserverJob = (event: any) => {
  return new PushImageJob(pushObserverJobName, event)
}
jobs[pushObserverJobName] = pushObserverJob

const buildSchedulerJobName = "build-scheduler"
const buildSchedulerJob = (event: any) => {
  return new MakeTargetJob(buildSchedulerJobName, kanikoImg, event)
}
jobs[buildSchedulerJobName] = buildSchedulerJob

const pushSchedulerJobName = "push-scheduler"
const pushSchedulerJob = (event: any) => {
  return new PushImageJob(pushSchedulerJobName, event)
}
jobs[pushSchedulerJobName] = pushSchedulerJob

const buildWorkerJobName = "build-worker"
const buildWorkerJob = (event: any) => {
  return new MakeTargetJob(buildWorkerJobName, kanikoImg, event)
}
jobs[buildWorkerJobName] = buildWorkerJob

const pushWorkerJobName = "push-worker"
const pushWorkerJob = (event: any) => {
  return new PushImageJob(pushWorkerJobName, event)
}
jobs[pushWorkerJobName] = pushWorkerJob

// Helm chart:

const lintChartJobName = "lint-chart"
const lintChartJob = (event: any) => {
  return new MakeTargetJob(lintChartJobName, helmImg, event)
}
jobs[lintChartJobName] = lintChartJob

const publishChartJobName = "publish-chart"
const publishChartJob = (event: any) => {
  return new MakeTargetJob(publishChartJobName, helmImg, event, {
    "HELM_REGISTRY": event.project.secrets.helmRegistry || "ghcr.io",
    "HELM_ORG": event.project.secrets.helmOrg,
    "HELM_USERNAME": event.project.secrets.helmUsername,
    "HELM_PASSWORD": event.project.secrets.helmPassword
  })
}
jobs[publishChartJobName] = publishChartJob

// CLI:

const buildCLIJobName = "build-cli"
const buildCLIJob = (event: any) => {
  return new MakeTargetJob(buildCLIJobName, goImg, event)
}
jobs[buildCLIJobName] = buildCLIJob

const publishCLIJobName = "publish-cli"
const publishCLIJob = (event: any) => {
  return new MakeTargetJob(publishCLIJobName, goImg, event, {
    "GITHUB_ORG": event.project.secrets.githubOrg,
    "GITHUB_REPO": event.project.secrets.githubRepo,
    "GITHUB_TOKEN": event.project.secrets.githubToken
  })
}
jobs[publishCLIJobName] = publishCLIJob

// Run the entire suite of tests, builds, etc. concurrently WITHOUT publishing
// anything initially. If EVERYTHING passes AND this was a push (merge,
// presumably) to the v2 branch, then run jobs to publish "edge" images.
async function runSuite(event: any): Promise<void> {
  await new ConcurrentGroup(
    // // Basic tests:
    // testUnitGoJob(event),
    // lintGoJob(event),

    // Use just these two for demo purposes-- they run pretty fast
    testUnitJSJob(event),
    lintJSJob(event)
    
    // // Brigadier:
    // buildBrigadierJob(event),
    // // Docker images:
    // buildAPIServerJob(event),
    // buildGitInitializerJob(event),
    // buildLoggerLinuxJob(event),
    // buildObserverJob(event),
    // buildSchedulerJob(event),
    // buildWorkerJob(event),
    // // Helm chart:
    // lintChartJob(event),
    // // CLI:
    // buildCLIJob(event)
  ).run()
  // if (event.revision.ref == "v2") {
  //   // Push "edge" images.
  //   //
  //   // npm packages MUST be semantically versioned, so we DON'T publish an
  //   // edge brigadier package.
  //   //
  //   // To keep our github released page tidy, we're also not publishing "edge"
  //   // CLI binaries.
  //   await new ConcurrentGroup(
  //     pushAPIServerJob(event),
  //     pushGitInitializerJob(event),
  //     pushLoggerLinuxJob(event),
  //     pushObserverJob(event),
  //     pushSchedulerJob(event),
  //     pushWorkerJob(event),
  //   ).run()
  // }
}

// Either of these events should initiate execution of the entire test suite.
events.on("github.com/krancour/brigade-github-gateway", "check_suite:requested", runSuite)
events.on("github.com/krancour/brigade-github-gateway", "check_suite:rerequested", runSuite)

// This event indicates a specific job is to be re-run.
events.on("github.com/krancour/brigade-github-gateway", "check_run:rerequested", async event => {
  const jobName = JSON.parse(event.payload).check_run.name
  const job = jobs[jobName]
  if (job) {
    await job(event).run()
  }
  throw new Error(`No job found with name: ${jobName}`)
})

events.process()
