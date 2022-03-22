import { events, Event, Job, ConcurrentGroup, SerialGroup, Container } from "@brigadecore/brigadier"

const azImg = "mcr.microsoft.com/azure-cli"
const goImg = "brigadecore/go-tools:v0.8.0"
const jsImg = "node:16.11.0-bullseye"
const dindImg = "docker:20.10.9-dind"
const dockerClientImg = "brigadecore/docker-tools:v0.2.0"
const helmImg = "brigadecore/helm-tools:v0.4.0"
const localPath = "/workspaces/brigade"

const dindSidecar = new Container(dindImg)
dindSidecar.privileged = true
dindSidecar.environment["DOCKER_TLS_CERTDIR"] = ""

// JobWithSource is a base class for any Job that uses project source code.
class JobWithSource extends Job {
  constructor(name: string, img: string, event: Event, env?: {[key: string]: string}) {
    super(name, img, event)
    this.primaryContainer.sourceMountPath = localPath
    this.primaryContainer.workingDirectory = localPath
    this.primaryContainer.environment = env || {}    
  }
}

// MakeTargetJob is just a job wrapper around one or more make targets.
class MakeTargetJob extends JobWithSource {
  constructor(name: string, targets: string[], img: string, event: Event, env?: {[key: string]: string}) {
    env ||= {}
    env["SKIP_DOCKER"] = "true"
    super(name, img, event, env)
    this.primaryContainer.command = [ "make" ]
    this.primaryContainer.arguments = targets
  }
}

// BuildImageJob is a specialized job type for building and pushing multiarch
// Docker images.
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
class BuildImageJob extends JobWithSource {
  constructor(image: string, event: Event, version?: string) {
    const secrets = event.project.secrets
    const env = {
      // Use the Docker daemon that's running in a sidecar
      "DOCKER_HOST": "localhost:2375"
    }
    let registry: string
    let registryOrg: string
    let registryUsername: string
    let registryPassword: string
    if (!version) { // This is where we'll push potentially unstable images
      registry = secrets.unstableImageRegistry
      registryOrg = secrets.unstableImageRegistryOrg
      registryUsername = secrets.unstableImageRegistryUsername
      registryPassword = secrets.unstableImageRegistryPassword
    } else { // This is where we'll push stable images only
      registry = secrets.stableImageRegistry
      registryOrg = secrets.stableImageRegistryOrg
      registryUsername = secrets.stableImageRegistryUsername
      registryPassword = secrets.stableImageRegistryPassword
      // Since it's defined, the make target will want this env var
      env["VERSION"] = version
    }
    if (registry) {
      // Since it's defined, the make target will want this env var
      env["DOCKER_REGISTRY"] = registry
    }
    if (registryOrg) {
      // Since it's defined, the make target will want this env var
      env["DOCKER_ORG"] = registryOrg
    }
    // We ALWAYS log in to Docker Hub because even if we plan to push the images
    // elsewhere, we still PULL a lot of images from Docker Hub (in FROM
    // directives of Dockerfiles) and we want to avoid being rate limited.
    env["DOCKERHUB_PASSWORD"] = secrets.dockerhubPassword
    let registriesLoginCmd = `docker login -u ${secrets.dockerhubUsername} -p $DOCKERHUB_PASSWORD`
    // If the registry we push to is defined (not DockerHub; which we're already
    // logging into) and we have credentials, add a second registry login:
    if (registry && registryUsername && registryPassword) {
      env["IMAGE_REGISTRY_PASSWORD"] = registryPassword
      registriesLoginCmd = `${registriesLoginCmd} && docker login ${registry} -u ${registryUsername} -p $IMAGE_REGISTRY_PASSWORD`
    }
    super(`build-${image}`, dockerClientImg, event, env)
    this.primaryContainer.command = [ "sh" ]
    this.primaryContainer.arguments = [
      "-c",
      // The sleep is a grace period after which we assume the DinD sidecar is
      // probably up and running.
      "sleep 20 && " +
        `${registriesLoginCmd} && ` +
        "docker buildx create --name builder --use && " +
        "docker info && " +
        `make push-${image}`
    ]
    this.sidecarContainers.dind = dindSidecar
  }
}

class ScanJob extends MakeTargetJob {
  constructor(image: string, event: Event) {
    const env = {}
    const secrets = event.project.secrets
    if (secrets.unstableImageRegistry) {
      env["DOCKER_REGISTRY"] = secrets.unstableImageRegistry
    }
    if (secrets.unstableImageRegistryOrg) {
      env["DOCKER_ORG"] = secrets.unstableImageRegistryOrg
    }
    super(`scan-${image}`, [`scan-${image}`], dockerClientImg, event, env)
    this.fallible = true
  }
}

class PublishSBOMJob extends MakeTargetJob {
  constructor(name: string, image: string, event: Event, version: string) {
    const secrets = event.project.secrets
    const env = {
      "GITHUB_ORG": secrets.githubOrg,
      "GITHUB_REPO": secrets.githubRepo,
      "GITHUB_TOKEN": secrets.githubToken,
      "VERSION": version
    }
    if (secrets.stableImageRegistry) {
      env["DOCKER_REGISTRY"] = secrets.stableImageRegistry
    }
    if (secrets.stableImageRegistryOrg) {
      env["DOCKER_ORG"] = secrets.stableImageRegistryOrg
    }
    super(name, [`publish-sbom-${image}`], dockerClientImg, event, env)
  }
}

// A map of all jobs. When a ci:job_requested event wants to re-run a single
// job, this allows us to easily find that job by name.
const jobs: {[key: string]: (event: Event, version?: string) => Job } = {}

// Basic tests:

const testUnitGoJobName = "test-unit-go"
const testUnitGoJob = (event: Event) => {
  return new MakeTargetJob(testUnitGoJobName, ["test-unit-go"], goImg, event)
}
jobs[testUnitGoJobName] = testUnitGoJob

const lintGoJobName = "lint-go"
const lintGoJob = (event: Event) => {
  return new MakeTargetJob(lintGoJobName, ["lint-go"], goImg, event)
}
jobs[lintGoJobName] = lintGoJob

const testUnitJSJobName = "test-unit-js"
const testUnitJSJob = (event: Event) => {
  return new MakeTargetJob(testUnitJSJobName, ["test-unit-js"], jsImg, event)
}
jobs[testUnitJSJobName] = testUnitJSJob

const styleCheckJSJobName = "style-check-js"
const styleCheckJSJob = (event: Event) => {
  return new MakeTargetJob(styleCheckJSJobName, ["style-check-js"], jsImg, event)
}
jobs[styleCheckJSJobName] = styleCheckJSJob

const lintJSJobName = "lint-js"
const lintJSJob = (event: Event) => {
  return new MakeTargetJob(lintJSJobName, ["lint-js"], jsImg, event)
}
jobs[lintJSJobName] = lintJSJob

const yarnAuditJobName = "yarn-audit"
const yarnAuditJob = (event: Event) => {
  const job = new MakeTargetJob(yarnAuditJobName, ["yarn-audit"], jsImg, event) 
  job.fallible = true
  return job
}
jobs[yarnAuditJobName] = yarnAuditJob

const lintChartJobName = "lint-chart"
const lintChartJob = (event: Event) => {
  return new MakeTargetJob(lintChartJobName, ["lint-chart"], helmImg, event)
}
jobs[lintChartJobName] = lintChartJob

const validateSchemasJobName = "validate-schemas"
const validateSchemasJob = (event: Event) => {
  return new MakeTargetJob(validateSchemasJobName, ["validate-schemas"], jsImg, event);
}
jobs[validateSchemasJobName] = validateSchemasJob

const validateExamplesJobName = "validate-examples"
const validateExamplesJob = (event: Event) => {
  return new MakeTargetJob(validateExamplesJobName, ["validate-examples"], jsImg, event)
}
jobs[validateExamplesJobName] = validateExamplesJob;

// Build / publish stuff:

const buildArtemisJobName = "build-artemis"
const buildArtemisJob = (event: Event, version?: string) => {
  return new BuildImageJob("artemis", event, version)
}
jobs[buildArtemisJobName] = buildArtemisJob

const scanArtemisJobName = "scan-artemis"
const scanArtemisJob = (event: Event) => {
  return new ScanJob("artemis", event)
}
jobs[scanArtemisJobName] = scanArtemisJob

const publishArtemisSBOMJobName = "publish-sbom-artemis"
const publishArtemisSBOMJob = (event: Event, version: string) => {
  return new PublishSBOMJob(publishArtemisSBOMJobName, "artemis", event, version)
}
jobs[publishArtemisSBOMJobName] = publishArtemisSBOMJob

const buildAPIServerJobName = "build-apiserver"
const buildAPIServerJob = (event: Event, version?: string) => {
  return new BuildImageJob("apiserver", event, version)
}
jobs[buildAPIServerJobName] = buildAPIServerJob

const scanAPIServerJobName = "scan-apiserver"
const scanAPIServerJob = (event: Event) => {
  return new ScanJob("apiserver", event)
}
jobs[scanAPIServerJobName] = scanAPIServerJob

const publishAPIServerSBOMJobName = "publish-sbom-apiserver"
const publishAPIServerSBOMJob = (event: Event, version: string) => {
  return new PublishSBOMJob(publishAPIServerSBOMJobName, "apiserver", event, version)
}
jobs[publishAPIServerSBOMJobName] = publishAPIServerSBOMJob

const buildGitInitializerJobName = "build-git-initializer"
const buildGitInitializerJob = (event: Event, version?: string) => {
  return new BuildImageJob("git-initializer", event, version)
}
jobs[buildGitInitializerJobName] = buildGitInitializerJob

const scanGitInitializerJobName = "scan-git-initializer"
const scanGitInitializerJob = (event: Event) => {
  return new ScanJob("git-initializer", event)
}
jobs[scanGitInitializerJobName] = scanGitInitializerJob

const publishGitInitializerSBOMJobName = "publish-sbom-git-initializer"
const publishGitInitializerSBOMJob = (event: Event, version: string) => {
  return new PublishSBOMJob(publishGitInitializerSBOMJobName, "git-initializer", event, version)
}
jobs[publishGitInitializerSBOMJobName] = publishGitInitializerSBOMJob

const buildLoggerLinuxJobName = "build-logger"
const buildLoggerLinuxJob = (event: Event, version?: string) => {
  return new BuildImageJob("logger", event, version)
}
jobs[buildLoggerLinuxJobName] = buildLoggerLinuxJob

const scanLoggerLinuxJobName = "scan-logger"
const scanLoggerLinuxJob = (event: Event) => {
  return new ScanJob("logger", event)
}
jobs[scanLoggerLinuxJobName] = scanLoggerLinuxJob

const publishLoggerSBOMJobName = "publish-sbom-logger"
const publishLoggerSBOMJob = (event: Event, version: string) => {
  return new PublishSBOMJob(publishLoggerSBOMJobName, "logger", event, version)
}
jobs[publishLoggerSBOMJobName] = publishLoggerSBOMJob

const buildObserverJobName = "build-observer"
const buildObserverJob = (event: Event, version?: string) => {
  return new BuildImageJob("observer", event, version)
}
jobs[buildObserverJobName] = buildObserverJob

const scanObserverJobName = "scan-observer"
const scanObserverJob = (event: Event) => {
  return new ScanJob("observer", event)
}
jobs[scanObserverJobName] = scanObserverJob

const publishObserverSBOMJobName = "publish-sbom-observer"
const publishObserverSBOMJob = (event: Event, version: string) => {
  return new PublishSBOMJob(publishObserverSBOMJobName, "observer", event, version)
}
jobs[publishObserverSBOMJobName] = publishObserverSBOMJob

const buildSchedulerJobName = "build-scheduler"
const buildSchedulerJob = (event: Event, version?: string) => {
  return new BuildImageJob("scheduler", event, version)
}
jobs[buildSchedulerJobName] = buildSchedulerJob

const scanSchedulerJobName = "scan-scheduler"
const scanSchedulerJob = (event: Event) => {
  return new ScanJob("scheduler", event)
}
jobs[scanSchedulerJobName] = scanSchedulerJob

const publishSchedulerSBOMJobName = "publish-sbom-scheduler"
const publishSchedulerSBOMJob = (event: Event, version: string) => {
  return new PublishSBOMJob(publishSchedulerSBOMJobName, "scheduler", event, version)
}
jobs[publishSchedulerSBOMJobName] = publishSchedulerSBOMJob

const buildWorkerJobName = "build-worker"
const buildWorkerJob = (event: Event, version?: string) => {
  return new BuildImageJob("worker", event, version)
}
jobs[buildWorkerJobName] = buildWorkerJob

const scanWorkerJobName = "scan-worker"
const scanWorkerJob = (event: Event) => {
  return new ScanJob("worker", event)
}
jobs[scanWorkerJobName] = scanWorkerJob

const publishWorkerSBOMJobName = "publish-sbom-worker"
const publishWorkerSBOMJob = (event: Event, version: string) => {
  return new PublishSBOMJob(publishWorkerSBOMJobName, "worker", event, version)
}
jobs[publishWorkerSBOMJobName] = publishWorkerSBOMJob

const buildBrigadierJobName = "build-brigadier"
const buildBrigadierJob = (event: Event) => {
  return new MakeTargetJob(buildBrigadierJobName, ["build-brigadier"], jsImg, event)
}
jobs[buildBrigadierJobName] = buildBrigadierJob

const publishBrigadierJobName = "publish-brigadier"
const publishBrigadierJob = (event: Event, version: string) => {
  return new MakeTargetJob(publishBrigadierJobName, ["publish-brigadier"], jsImg, event, {
    "VERSION": version,
    "NPM_TOKEN": event.project.secrets.npmToken
  })
}
jobs[publishBrigadierJobName] = publishBrigadierJob

const buildCLIJobName = "build-cli"
const buildCLIJob = (event: Event) => {
  return new MakeTargetJob(buildCLIJobName, ["build-cli"], goImg, event)
}
jobs[buildCLIJobName] = buildCLIJob

const publishCLIJobName = "publish-cli"
const publishCLIJob = (event: Event, version: string) => {
  const secrets = event.project.secrets
  return new MakeTargetJob(publishCLIJobName, ["publish-cli"], goImg, event, {
    "VERSION": version,
    "GITHUB_ORG": secrets.githubOrg,
    "GITHUB_REPO": secrets.githubRepo,
    "GITHUB_TOKEN": secrets.githubToken
  })
}
jobs[publishCLIJobName] = publishCLIJob

const publishChartJobName = "publish-chart"
const publishChartJob = (event: Event, version: string) => {
  const secrets = event.project.secrets
  const helmRegistry = secrets.chartRegistry || "ghcr.io"
  const job = new MakeTargetJob(publishChartJobName, ["publish-chart"], helmImg, event, {
    "VERSION": version,
    "HELM_REGISTRY": helmRegistry,
    "HELM_ORG": secrets.helmOrg,
    "HELM_REGISTRY_PASSWORD": secrets.helmPassword
  })
  job.primaryContainer.command = [ "sh" ]
  job.primaryContainer.arguments = [
    "-c",
    `helm registry login ${helmRegistry} -u ${secrets.helmUsername} -p $HELM_REGISTRY_PASSWORD && ` +
      "make publish-chart"
  ]
  return job
}
jobs[publishChartJobName] = publishChartJob

const publishBrigadierDocsJobName = "publish-brigadier-docs"
const publishBrigadierDocsJob = (event: Event, version?: string) => {
  return new MakeTargetJob(publishBrigadierDocsJobName, ["publish-brigadier-docs"], jsImg, event, {
    "VERSION": version,
    "GH_TOKEN": event.project.secrets.ghToken
  })
}
jobs[publishBrigadierDocsJobName] = publishBrigadierDocsJob

const testIntegrationJobName = "test-integration"
const testIntegrationJob = (event: Event) => {
  const secrets = event.project.secrets
  const env = {
    "CGO_ENABLED": "0",
    "BRIGADE_CI_PRIVATE_REPO_SSH_KEY": secrets.privateRepoSSHKey,
    "IMAGE_PULL_POLICY": "IfNotPresent",
    // Use the Docker daemon that's running in a sidecar
    "DOCKER_HOST": "localhost:2375"
  }
  if (secrets.unstableImageRegistry) {
    env["DOCKER_REGISTRY"] = secrets.unstableImageRegistry
  }
  if (secrets.unstableImageRegistryOrg) {
    env["DOCKER_ORG"] = secrets.unstableImageRegistryOrg
  }
  const job = new MakeTargetJob(testIntegrationJobName, ["test-integration"], "brigadecore/int-test-tools:v0.2.0", event, env)
  job.primaryContainer.command = [ "sh" ]
  job.primaryContainer.arguments = [
    "-c",
    // The sleep is a grace period after which we assume the DinD sidecar is
    // probably up and running.
    "sleep 20 && " +
      "docker info && " +
      "kind create cluster && " +
      "make hack-deploy test-integration"
  ]
  job.sidecarContainers.dind = dindSidecar
  job.timeoutSeconds = 30 * 60
  return job
}
jobs[testIntegrationJobName] = testIntegrationJob

events.on("brigade.sh/github", "ci:pipeline_requested", async event => {
  await new SerialGroup(
    new ConcurrentGroup( // Basic tests
      testUnitGoJob(event),
      lintGoJob(event),
      testUnitJSJob(event),
      styleCheckJSJob(event),
      lintJSJob(event),
      yarnAuditJob(event),
      lintChartJob(event),
      validateSchemasJob(event),
      validateExamplesJob(event)
    ),
    new ConcurrentGroup( // Build everything
      new SerialGroup(buildArtemisJob(event), scanArtemisJob(event)),
      new SerialGroup(buildAPIServerJob(event), scanAPIServerJob(event)),
      new SerialGroup(buildGitInitializerJob(event), scanGitInitializerJob(event)),
      new SerialGroup(buildLoggerLinuxJob(event), scanLoggerLinuxJob(event)),
      new SerialGroup(buildObserverJob(event), scanObserverJob(event)),
      new SerialGroup(buildSchedulerJob(event), scanSchedulerJob(event)),
      new SerialGroup(buildWorkerJob(event), scanWorkerJob(event)),
      buildBrigadierJob(event),
      buildCLIJob(event)
    ),
    testIntegrationJob(event)
  ).run()
})

// This event indicates a specific job is to be re-run.
events.on("brigade.sh/github", "ci:job_requested", async event => {
  const job = jobs[event.labels.job]
  if (job) {
    await job(event).run()
    return
  }
  throw new Error(`No job found with name: ${event.labels.job}`)
})

events.on("brigade.sh/github", "cd:pipeline_requested", async event => {
  const version = JSON.parse(event.payload).release.tag_name
  await new SerialGroup(
    new ConcurrentGroup(
      new SerialGroup(buildArtemisJob(event, version), publishArtemisSBOMJob(event, version)),
      new SerialGroup(buildAPIServerJob(event, version), publishAPIServerSBOMJob(event, version)),
      new SerialGroup(buildGitInitializerJob(event, version), publishGitInitializerSBOMJob(event, version)),
      new SerialGroup(buildLoggerLinuxJob(event, version), publishLoggerSBOMJob(event, version)),
      new SerialGroup(buildObserverJob(event, version), publishObserverSBOMJob(event, version)),
      new SerialGroup(buildSchedulerJob(event, version), publishSchedulerSBOMJob(event, version)),
      new SerialGroup(buildWorkerJob(event, version), publishWorkerSBOMJob(event, version))
    ),
    new ConcurrentGroup(
      publishBrigadierJob(event, version),
      publishBrigadierDocsJob(event, version),
      publishChartJob(event, version),
      publishCLIJob(event, version)
    )
  ).run()
})

events.on("brigade.sh/cron", "nightly-cleanup", async event => {
  const secrets = event.project.secrets
  const job = new Job("unstable-acr-cleanup", azImg, event)
  job.primaryContainer.environment = {
    "AZ_PASSWORD": secrets.azPassword
  }
  job.primaryContainer.command = ["sh"]
  let script = `az login --service-principal --username ${secrets.azUsername} --password $AZ_PASSWORD --tenant ${secrets.azTenant} `
  const repos = [
    "brigade2-apiserver",
    "brigade2-artemis",
    "brigade2-git-initializer",
    "brigade2-logger",
    "brigade2-observer",
    "brigade2-scheduler",
    "brigade2-worker"
  ]
  repos.forEach((repo: string) => {
    script += `&& az acr repository delete --name unstablebrigade --repository ${repo} --yes`
  })
  job.primaryContainer.arguments = ["-c", script]
  await job.run()
})

events.process()
