import { events, Event, Job, ConcurrentGroup, SerialGroup, Container } from "@brigadecore/brigadier"

const releaseStoragePath = "/release-assets"

const goImg = "brigadecore/go-tools:v0.1.0"
const jsImg = "node:12.22.7-bullseye"
const dindImg = "docker:20.10.9-dind"
const dockerClientImg = "brigadecore/docker-tools:v0.1.0"

const projectName = "brigade"
const projectOrg = "brigadecore"
const gopath = "/go"
const localPath = gopath + `/src/github.com/${projectOrg}/${projectName}`

// A map of all jobs. When a check_run:rerequested event wants to re-run a
// single job, this allows us to easily find that job by name.
const jobs: {[key: string]: (event: Event, version?: string) => Job } = {}

// FallibleJob is a Job that is allowed to fail without compelling the worker
// process to fail.
//
// TODO: This will no longer be needed after
// https://github.com/brigadecore/brigade/issues/1768 is addressed.
class FallibleJob extends Job {
  constructor(name: string, image: string, event: Event) {
    super(name, image, event)
  }
  async run(): Promise<void> {
    try {
      await super.run()
    } catch {
      // Deliberately sweep any error under the rug
    }
    return Promise.resolve()
  }
}

const testGoJobName = "test-go"
const testGoJob = (event: Event) => {
  const job = new Job(testGoJobName, goImg, event)
  job.primaryContainer.sourceMountPath = localPath
  job.primaryContainer.workingDirectory = localPath
  job.primaryContainer.environment = {
    "SKIP_DOCKER": "true"
  }
  job.primaryContainer.command = [ "make" ]
  job.primaryContainer.arguments = [ "verify-vendored-code", "lint", "test-unit" ]
  return job
}
jobs[testGoJobName] = testGoJob

const testJSJobName = "test-javascript"
const testJSJob = (event: Event) => {
  const job = new Job(testJSJobName, jsImg, event)
  job.primaryContainer.sourceMountPath = localPath
  job.primaryContainer.workingDirectory = localPath
  job.primaryContainer.environment = {
    "SKIP_DOCKER": "true"
  }
  job.primaryContainer.command = [ "make" ]
  job.primaryContainer.arguments = [ "test-js" ]
  return job
}
jobs[testJSJobName] = testJSJob

const yarnAuditJobName = "yarn-audit"
const yarnAuditJob = (event: Event) => {
  const job = new FallibleJob(yarnAuditJobName, jsImg, event)
  job.primaryContainer.sourceMountPath = localPath
  job.primaryContainer.workingDirectory = localPath
  job.primaryContainer.environment = {
    "SKIP_DOCKER": "true"
  }
  job.primaryContainer.command = [ "make" ]
  job.primaryContainer.arguments = [ "yarn-audit" ]
  return job
}
jobs[yarnAuditJobName] = yarnAuditJob

const e2eJobName = "test-e2e"
const e2eJob = (event: Event) => {
  const dockerOrg = event.project.secrets.dockerhubOrg || "brigadecore"
  const job = new Job(e2eJobName, "brigadecore/int-test-tools:v0.2.0", event)
  job.primaryContainer.sourceMountPath = localPath
  job.primaryContainer.workingDirectory = localPath
  job.primaryContainer.environment = {
    "DOCKER_HOST": "localhost:2375",
    "DOCKER_ORG": dockerOrg
  }
  job.primaryContainer.command = [ "sh" ]
  job.primaryContainer.arguments = [
    "-c",
    // The sleep is a grace period after which we assume the DinD sidecar is
    // probably up and running.
    `sleep 20 && kind create cluster && helm repo add brigade https://brigadecore.github.io/charts && make build-all-images load-all-images helm-install e2e`
  ]
  job.sidecarContainers.docker = new Container(dindImg)
  job.sidecarContainers.docker.privileged = true
  job.sidecarContainers.docker.environment.DOCKER_TLS_CERTDIR=""
  job.timeoutSeconds = 60 * 60
  return job
}
jobs[e2eJobName] = e2eJob

const buildAndPublishImagesJobName = "build-and-publish-images"
const buildAndPublishImagesJob = (event: Event, version?: string) => {
  const dockerRegistry = event.project.secrets.dockerhubRegistry || "docker.io"
  const dockerOrg = event.project.secrets.dockerhubOrg || "brigadecore"
  const job = new Job(buildAndPublishImagesJobName, dockerClientImg, event)
  job.primaryContainer.sourceMountPath = localPath
  job.primaryContainer.workingDirectory = localPath
  job.primaryContainer.environment = {
    "DOCKER_HOST": "localhost:2375",
    "DOCKER_REGISTRY": dockerRegistry,
    "DOCKER_ORG": dockerOrg,
    "DOCKER_PASSWORD": event.project.secrets.dockerhubPassword
  }
  if (version) {
    job.primaryContainer.environment["VERSION"] = version
  }
  job.primaryContainer.command = [ "sh" ]
  job.primaryContainer.arguments = [
    "-c",
    // The sleep is a grace period after which we assume the DinD sidecar is
    // probably up and running.
    `sleep 20 && docker login ${dockerRegistry} -u ${event.project.secrets.dockerhubUsername} -p $DOCKER_PASSWORD && make build-all-images push-all-images`
  ]
  job.sidecarContainers.docker = new Container(dindImg)
  job.sidecarContainers.docker.privileged = true
  job.sidecarContainers.docker.environment.DOCKER_TLS_CERTDIR=""
  return job
}
jobs[buildAndPublishImagesJobName] = buildAndPublishImagesJob

const buildBrigJobName = "build-brig"
const buildBrigJob = (event: Event, version?: string) => {
  const job = new Job(buildBrigJobName, goImg, event)
  job.primaryContainer.sourceMountPath = localPath
  job.primaryContainer.workingDirectory = localPath
  job.primaryContainer.workspaceMountPath = releaseStoragePath
  job.primaryContainer.environment = {
    "SKIP_DOCKER": "true",
  }
  if (version) {
    job.primaryContainer.environment["VERSION"] = version
  }
  job.primaryContainer.arguments = [
    "-c",
    `make xbuild-brig && cp -r bin/* ${releaseStoragePath}`
  ]
  return job
}
jobs[buildBrigJobName] = buildBrigJob

const githubReleaseJobName = "github-release"
const githubReleaseJob = (event: Event, version: string) => {
  const job = new Job(githubReleaseJobName, goImg, event)
  job.primaryContainer.sourceMountPath = localPath
  job.primaryContainer.workingDirectory = localPath
  job.primaryContainer.workspaceMountPath = releaseStoragePath
  job.primaryContainer.environment = {
    "SKIP_DOCKER": "true",
    "GITHUB_ORG": event.project.secrets.githubOrg,
    "GITHUB_REPO": event.project.secrets.githubRepo,
    "GITHUB_TOKEN": event.project.secrets.githubToken,
    "VERSION": version
  }
  job.primaryContainer.arguments = [
    "-c",
    `go get github.com/tcnksm/ghr && ghr -u $GITHUB_ORG -r $GITHUB_REPO -c $(git rev-parse HEAD) -t $GITHUB_TOKEN -n $VERSION $VERSION ${releaseStoragePath}`
  ]
  return job
}
jobs[githubReleaseJobName] = githubReleaseJob

// Run the entire suite of tests WITHOUT publishing anything initially. If
// EVERYTHING passes AND this was a push (merge, presumably) to the v1 branch,
// then run jobs to publish "edge" images.
async function runSuite(event: Event): Promise<void> {
  await new ConcurrentGroup( // Basic tests
    testGoJob(event),
    testJSJob(event),
    yarnAuditJob(event),
    e2eJob(event)
  ).run()
  if (event.worker?.git?.ref == "v1") {
    await buildAndPublishImagesJob(event).run()
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

events.on("brigade.sh/github", "release:published", async event => {
  const version = JSON.parse(event.payload).release.tag_name
  await new SerialGroup(
    new ConcurrentGroup(
      testGoJob(event),
      testJSJob(event),
      yarnAuditJob(event),
      e2eJob(event)
    ),
    buildAndPublishImagesJob(event, version),
    buildBrigJob(event, version),
    githubReleaseJob(event, version)
  ).run()
})

events.process()
