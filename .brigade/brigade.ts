import { events, Event, Job, ConcurrentGroup, SerialGroup, Container } from "@brigadecore/brigadier"

const releaseTagRegex = /^refs\/tags\/(v[0-9]+(?:\.[0-9]+)*(?:\-.+)?)$/

const goImg = "brigadecore/go-tools:v0.5.0"
const jsImg = "node:16.11.0-bullseye"
const dindImg = "docker:20.10.9-dind"
const dockerClientImg = "brigadecore/docker-tools:v0.1.0"
const helmImg = "brigadecore/helm-tools:v0.4.0"
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

// BuildImageJob is a specialized job type for building multiarch Docker images.
//
// Note: This isn't the optimal way to do this. It's a workaround. These notes
// are here so that as the situation improves, we can improve our approach.
//
// The optimal way of doing this would involve no sidecars and wouldn't closely
// resemble the "DinD" (Docker in Docker) pattern that we are accustomed to.
//
// `docker buildx build` has full support for building images using remote
// BuildKit instances. Such instances can use qemu to emulate other CPU
// architectures. This permits us to build images for arm64 (aka arm64/v8, aka
// aarch64), even though, as of this writing, we only have access to amd64 VMs.
//
// In an ideal world, we'd have a pool of BuildKit instances up and running at
// all times in our cluster and we'd somehow JOIN it and be off to the races.
// Alas, as of this writing, this isn't supported yet. (BuildKit supports it,
// but the `docker buildx` family of commands does not.) The best we can do is
// use `docker buildx create` to create a brand new builder.
//
// Tempting as it is to create a new builder using the Kubernetes driver (i.e.
// `docker buildx create --driver kubernetes`), this comes with two problems:
// 
// 1. It would require giving our jobs a lot of additional permissions that they
//    don't otherwise need (creating deployments, for instance). This represents
//    an attack vector I'd rather not open.
//
// 2. If the build should fail, nothing guarantees the builder gets shut down.
//    Over time, this could really clutter the cluster and starve us of
//    resources.
//
// The workaround I have chosen is to launch a new builder using the default
// docker-container driver. This runs inside a DinD sidecar. This has the
// benefit of always being cleaned up when the job is observed complete by the
// Brigade observer. The downside is that we're building an image inside a
// Russian nesting doll of containers with an ephemeral cache. It is slow, but
// it works.
//
// If and when the capability exists to use `docker buildx` with existing
// builders, we can streamline all of this pretty significantly.
class BuildImageJob extends MakeTargetJob {
  constructor(target: string, event: Event, env?: {[key: string]: string}) {
    super(target, dockerClientImg, event, env)
    this.primaryContainer.environment.DOCKER_HOST = "localhost:2375"
    this.primaryContainer.command = [ "sh" ]
    this.primaryContainer.arguments = [
      "-c",
      // The sleep is a grace period after which we assume the DinD sidecar is
      // probably up and running.
      `sleep 20 && docker buildx create --name builder --use && docker buildx ls && make ${target}`
    ]

    this.sidecarContainers.docker = new Container(dindImg)
    this.sidecarContainers.docker.privileged = true
    this.sidecarContainers.docker.environment.DOCKER_TLS_CERTDIR=""
  }
}

// PushImageJob is a specialized job type for publishing Docker images.
class PushImageJob extends BuildImageJob {
  constructor(target: string, event: Event) {
    super(target, event, {
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

const buildArtemisJobName = "build-artemis"
const buildArtemisJob = (event: Event) => {
  return new BuildImageJob(buildArtemisJobName, event)
}
jobs[buildArtemisJobName] = buildArtemisJob

const pushArtemisJobName = "push-artemis"
const pushArtemisJob = (event: Event) => {
  return new PushImageJob(pushArtemisJobName, event)
}
jobs[pushArtemisJobName] = pushArtemisJob

const buildAPIServerJobName = "build-apiserver"
const buildAPIServerJob = (event: Event) => {
  return new BuildImageJob(buildAPIServerJobName, event)
}
jobs[buildAPIServerJobName] = buildAPIServerJob

const pushAPIServerJobName = "push-apiserver"
const pushAPIServerJob = (event: Event) => {
  return new PushImageJob(pushAPIServerJobName, event)
}
jobs[pushAPIServerJobName] = pushAPIServerJob

const buildGitInitializerJobName = "build-git-initializer"
const buildGitInitializerJob = (event: Event) => {
  return new BuildImageJob(buildGitInitializerJobName, event)
}
jobs[buildGitInitializerJobName] = buildGitInitializerJob

const pushGitInitializerJobName = "push-git-initializer"
const pushGitInitializerJob = (event: Event) => {
  return new PushImageJob(pushGitInitializerJobName, event)
}
jobs[pushGitInitializerJobName] = pushGitInitializerJob

const buildLoggerLinuxJobName = "build-logger-linux"
const buildLoggerLinuxJob = (event: Event) => {
  return new BuildImageJob(buildLoggerLinuxJobName, event)
}
jobs[buildLoggerLinuxJobName] = buildLoggerLinuxJob

const pushLoggerLinuxJobName = "push-logger-linux"
const pushLoggerLinuxJob = (event: Event) => {
  return new PushImageJob(pushLoggerLinuxJobName, event)
}
jobs[pushLoggerLinuxJobName] = pushLoggerLinuxJob

const buildObserverJobName = "build-observer"
const buildObserverJob = (event: Event) => {
  return new BuildImageJob(buildObserverJobName, event)
}
jobs[buildObserverJobName] = buildObserverJob

const pushObserverJobName = "push-observer"
const pushObserverJob = (event: Event) => {
  return new PushImageJob(pushObserverJobName, event)
}
jobs[pushObserverJobName] = pushObserverJob

const buildSchedulerJobName = "build-scheduler"
const buildSchedulerJob = (event: Event) => {
  return new BuildImageJob(buildSchedulerJobName, event)
}
jobs[buildSchedulerJobName] = buildSchedulerJob

const pushSchedulerJobName = "push-scheduler"
const pushSchedulerJob = (event: Event) => {
  return new PushImageJob(pushSchedulerJobName, event)
}
jobs[pushSchedulerJobName] = pushSchedulerJob

const buildWorkerJobName = "build-worker"
const buildWorkerJob = (event: Event) => {
  return new BuildImageJob(buildWorkerJobName, event)
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

const publishBrigadierDocsJobName = "publish-brigadier-docs"
const publishBrigadierDocsJob = (event: Event) => {
  return new MakeTargetJob(publishBrigadierDocsJobName, jsImg, event, {
    "GH_TOKEN": event.project.secrets.ghToken
  })
}
jobs[publishBrigadierDocsJobName] = publishBrigadierDocsJob

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
      buildArtemisJob(event),
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
    // To keep our github releases page tidy, we're also not publishing "edge"
    // CLI binaries.
    await new ConcurrentGroup(
      pushArtemisJob(event),
      pushAPIServerJob(event),
      pushGitInitializerJob(event),
      pushLoggerLinuxJob(event),
      pushObserverJob(event),
      pushSchedulerJob(event),
      pushWorkerJob(event),
      publishBrigadierDocsJob(event),
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
        pushArtemisJob(event),
        pushAPIServerJob(event),
        pushGitInitializerJob(event),
        pushLoggerLinuxJob(event),
        pushObserverJob(event),
        pushSchedulerJob(event),
        pushWorkerJob(event)
      ),
      new ConcurrentGroup(
        publishBrigadierJob(event),
        publishBrigadierDocsJob(event),
        publishChartJob(event),
        publishCLIJob(event)
      )
    ).run()
  } else {
    console.log(`Ref ${event.worker.git.ref} does not match release tag regex (${releaseTagRegex}); not releasing.`)
  }
})

events.process()
