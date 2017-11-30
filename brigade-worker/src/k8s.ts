/**
 * k8s contains the Kubernetes implementation of Brigade.
 */

/** */

import * as kubernetes from '@kubernetes/typescript-node'
import * as jobs from './job'
import {BrigadeEvent, Project} from './events'
import { hostname } from 'os';

// The internals for running tasks. This must be loaded before any of the
// objects that use run().
//
// All Kubernetes API calls should be localized here. Other modules should not
// call 'kubernetes' directly.

// expiresInMSec is the number of milliseconds until pod expiration
// After this point, the pod can be garbage collected (a feature not yet implemented)
const expiresInMSec = 1000 * 60 * 60 * 24 * 30

const defaultClient = kubernetes.Config.defaultClient()

const serviceAccount = "brigade-worker"

class K8sResult implements jobs.Result {
  data: string
  constructor(msg: string) { this.data = msg }
  toString(): string { return this.data }
}

/**
 * BuildStorage manages per-build storage for a build.
 *
 * BuildStorage implements the app.BuildStorage interface.
 *
 * Storage is implemented as a PVC. The PVC backing it MUST be ReadWriteMany.
 */
export class BuildStorage {
  proj: Project
  name: string
  build: string

  /**
   * create initializes a new PVC for storing data.
   */
  public create(e: BrigadeEvent, project: Project, size: string): Promise<string> {
    this.proj = project
    this.name = e.workerID.toLowerCase()
    this.build = e.buildID
    let pvc = this.buildPVC(size)
    console.log(`Creating PVC named ${ this.name }`)
    return defaultClient.createNamespacedPersistentVolumeClaim(this.proj.kubernetes.namespace, pvc)
      .then( () => {return this.name })
  }
  /**
   * destroy deletes the PVC.
   */
  public destroy(): Promise<boolean> {
    console.log("Destroying PVC named " + this.name)
    let opts = new kubernetes.V1DeleteOptions()
    return defaultClient.deleteNamespacedPersistentVolumeClaim(this.name, this.proj.kubernetes.namespace, opts)
      .then( () => { return true })
  }
  /**
   * Get a PVC for a volume that lives for the duration of a build.
   */
  protected buildPVC(size: string): kubernetes.V1PersistentVolumeClaim {
    let s = new kubernetes.V1PersistentVolumeClaim()
    s.metadata = new kubernetes.V1ObjectMeta()
    s.metadata.name = this.name
    s.metadata.labels = {
      "heritage": "brigade",
      "component": "buildStorage",
      "project": this.proj.id,
      "worker": this.name,
      "build": this.build,
    }

    s.spec = new kubernetes.V1PersistentVolumeClaimSpec()
    s.spec.accessModes = ["ReadWriteMany"]

    let res = new kubernetes.V1ResourceRequirements()
    res.requests = { storage: size }
    s.spec.resources = res

    return s
  }
}

/**
 * loadProject takes a Secret name and namespace and loads the Project
 * from the secret.
 */
export function loadProject(name: string, ns: string): Promise<Project> {
  return defaultClient.readNamespacedSecret(name, ns).then( result => {
    return secretToProject(ns, result.body)
  })
}

/**
 * JobRunner provides a Kubernetes implementation of the JobRunner interface.
 */
export class JobRunner implements jobs.JobRunner {

  name: string
  secret: kubernetes.V1Secret
  runner: kubernetes.V1Pod
  pvc: kubernetes.V1PersistentVolumeClaim
  project: Project
  event: BrigadeEvent
  job: jobs.Job
  client: kubernetes.Core_v1Api

  constructor(job: jobs.Job, e: BrigadeEvent, project: Project) {
    this.event = e
    this.job = job
    this.project  = project
    this.client = defaultClient

    // $JOB-$TIME-$GITSHA
    let commit = e.commit || "master"
    this.name = job.name + "-" + Date.now() + "-" + commit.substring(0, 8);
    let secName = this.name
    let runnerName = this.name

    this.secret = newSecret(secName)
    this.runner = newRunnerPod(runnerName, job.image, job.imageForcePull)

    // Experimenting with setting a deadline field after which something
    // can clean up existing builds.
    let expiresAt = Date.now() + expiresInMSec

    this.runner.metadata.labels.jobname = job.name
    this.runner.metadata.labels.project = project.id
    this.runner.metadata.labels.commit = commit
    this.runner.metadata.labels.worker = e.workerID
    this.runner.metadata.labels.build = e.buildID

    this.secret.metadata.labels.jobname = job.name
    this.secret.metadata.labels.project = project.id
    this.secret.metadata.labels.commit = commit
    this.secret.metadata.labels.expires = String(expiresAt)
    this.secret.metadata.labels.worker = e.workerID
    this.secret.metadata.labels.build = e.buildID

    let envVars: kubernetes.V1EnvVar[] = []
    for (let key in job.env) {
      let val = job.env[key]
      this.secret.data[key] = b64enc(val)

      // Add reference to pod
      envVars.push({
        name: key,
        valueFrom: {
          secretKeyRef: {
            name: secName,
            key: key
          }
        }
      } as kubernetes.V1EnvVar)
    }


    // Do we still want to add this to the image directly? While it is a security
    // thing, not adding it would result in users not being able to push anything
    // upstream into the pod.
    if (project.repo.sshKey) {
      this.secret.data.brigadeSSHKey = b64enc(project.repo.sshKey)
      envVars.push({
        name: "BRIGADE_REPO_KEY",
        valueFrom: {
          secretKeyRef: {
            key: "brigadeSSHKey",
            name: secName
          }
        }
      } as kubernetes.V1EnvVar)
    }

    // Add top-level env vars. These must override any attempt to set the values
    // to something else.
    envVars.push(envVar("CLONE_URL", project.repo.cloneURL))
    envVars.push(envVar("HEAD_COMMIT_ID", e.commit))
    envVars.push(envVar("CI", "true"))

    this.runner.spec.containers[0].env = envVars

    let mountPath = job.mountPath || "/src"

    // Add secret volume
    this.runner.spec.volumes = [
      { name: secName, secret: {secretName: secName }} as kubernetes.V1Volume,
      { name: "vcs-sidecar", emptyDir: {}} as kubernetes.V1Volume
    ];
    this.runner.spec.containers[0].volumeMounts = [
      { name: secName, mountPath: "/hook"} as kubernetes.V1VolumeMount,
      { name: "vcs-sidecar", mountPath: mountPath} as kubernetes.V1VolumeMount
    ];

    if (job.useSource && project.repo.cloneURL) {
      // Add the sidecar.
      let sidecar = sidecarSpec(e, "/src", project.kubernetes.vcsSidecar, project, secName)
      this.runner.spec.initContainers = [sidecar]
    }

    if (job.imagePullSecrets) {
      this.runner.spec.imagePullSecrets = []
      for (let secret of job.imagePullSecrets) {
        this.runner.spec.imagePullSecrets.push({name: secret})
      }
    }

    // If host os is set, specify it.
    if (job.host.os) {
      this.runner.spec.nodeSelector = {
        "beta.kubernetes.io/os": job.host.os
      }
    }
    if (job.host.name) {
      this.runner.spec.nodeName = job.host.name
    }

    // If the job requests a cache, set up the cache.
    if (job.cache.enabled) {
      this.pvc = this.cachePVC()

      // Now add volume mount to pod:
      let mountName = this.cacheName()
      this.runner.spec.volumes.push({
        name: mountName,
        persistentVolumeClaim: {claimName: mountName}
      } as kubernetes.V1Volume )
      let mnt = volumeMount(mountName, job.cache.path)
      this.runner.spec.containers[0].volumeMounts.push(mnt)
    }

    // If the job needs build-wide storage, enable it.
    if (job.storage.enabled) {
      const vname = "build-storage"
      this.runner.spec.volumes.push({
        name: vname,
        persistentVolumeClaim: {claimName: e.buildID.toLowerCase()}
      } as kubernetes.V1Volume )
      let mnt = volumeMount(vname, job.storage.path)
      this.runner.spec.containers[0].volumeMounts.push(mnt)
    }

    // If the job needs access to a docker daemon, mount in the host's docker socket
    if (job.docker.enabled && project.allowHostMounts) {
      var dockerVol = new kubernetes.V1Volume()
      var dockerMount = new kubernetes.V1VolumeMount()
      var hostPath = new kubernetes.V1HostPathVolumeSource()
      hostPath.path = jobs.dockerSocketMountPath
      dockerVol.name = jobs.dockerSocketMountName
      dockerVol.hostPath = hostPath
      dockerMount.name = jobs.dockerSocketMountName
      dockerMount.mountPath = jobs.dockerSocketMountPath
      this.runner.spec.volumes.push(dockerVol)
      for (let i = 0; i < this.runner.spec.containers.length; i++) {
        this.runner.spec.containers[i].volumeMounts.push(dockerMount)
      }
    }


    let newCmd = generateScript(job)
    if (!newCmd) {
      this.runner.spec.containers[0].command = null
    } else {
      this.secret.data["main.sh"] = b64enc(newCmd)
    }

    // If the job askes for privileged mode and the project allows this, enable it.
    if (job.privileged && project.allowPrivilegedJobs) {
      for (let i = 0; i < this.runner.spec.containers.length; i++) {
        this.runner.spec.containers[i].securityContext.privileged = true
      }
    }

  }

  /**
   * cacheName returns the name of this job's cache PVC.
   */
  protected cacheName(): string {
    // The Kubernetes rules on pvc names are stupid^b^b^b^b strict. Name must
    // be DNS-like, and less than 64 chars. This rules out using project ID,
    // project name, etc. For now, we use project name with slashes replaced,
    // appended to job name.
    return `${ this.project.name.replace(/[.\/]/g, "-")}-${ this.job.name }`
  }

  /**
   * run starts a job and then waits until it is running.
   *
   * The Promise it returns will return when the pod is either marked
   * Success (resolve) or Failure (reject)
   */
  public run(): Promise<jobs.Result> {
    let podName = this.name
    let k = this.client
    let ns = this.project.kubernetes.namespace
    return this.start().then(r => r.wait()).then( r => {
      return k.readNamespacedPodLog(podName, ns)
    }).then(response => {
      return new K8sResult(response.body)
    })
  }

  /** start begins a job, and returns once it is scheduled to run.*/
  public start(): Promise<jobs.JobRunner> {
    // Now we have pod and a secret defined. Time to create them.

    let ns = this.project.kubernetes.namespace
    let k = this.client
    let pvcPromise = this.checkOrCreateCache()

    return new Promise((resolve, reject) => {
      pvcPromise.then( () => {
        console.log("Creating secret " + this.secret.metadata.name)
        return k.createNamespacedSecret(ns, this.secret)
      }).then((result) => {
        console.log("Creating pod " + this.runner.metadata.name)
        // Once namespace creation has been accepted, we create the pod.
        return k.createNamespacedPod(ns, this.runner)
      }).then((result) => {
          resolve(this)
      }).catch(reason => {
        if (reason.body) {
          console.log(reason.body)
        }
        reject(reason)
      })
    })
  }

  /**
   * checkOrCreateCache handles creating the cache if necessary.
   *
   * If no cache is requested by the job, this is a no-op.
   *
   * Otherwise, this checks for a cache, and if not found, it creates one.
   */
  protected checkOrCreateCache(): Promise<string> {
    return new Promise((resolve, reject) => {
      let ns = this.project.kubernetes.namespace
      let k = this.client
      if (!this.pvc) {
        resolve("no cache requested")
      }

      let cname = this.cacheName()
      console.log(`looking up ${ ns }/${ cname }`)
      k.readNamespacedPersistentVolumeClaim(cname, ns).then( result => {
        resolve("re-using existing cache")
      }).catch( result => {
        // TODO: check if cache exists.
        console.log("Creating Job Cache PVC " + cname)
        return k.createNamespacedPersistentVolumeClaim(ns,this.pvc).then((result, newPVC) => {
          console.log("created cache")
          resolve("created job cache")
        })
      }).catch( err => {
        console.error(err.body)
        reject(err)
      })
    })
  }

  /** wait listens for the running job to complete.*/
  public wait(): Promise<jobs.Result> {
    // Should probably protect against the case where start() was not called
    let k = this.client
    let timeout = this.job.timeout || 60000
    let name = this.name
    let ns = this.project.kubernetes.namespace
    let cancel = false

    // This is a handle to clear the setTimeout when the promise is fulfilled.
    let waiter

    console.log("Timeout set at " + timeout)

    // At intervals, poll the Kubernetes server and get the pod phase. If the
    // phase is Succeeded or Failed, bail out. Otherwise, keep polling.
    //
    // The timeout sets an upper limit, and if that limit is reached, the
    // polling will be stopped.
    //
    // Essentially, we track two Timer objects: the setTimeout and the setInterval.
    // That means we have to kill both before exit, otherwise the node.js process
    // will remain running until all timeouts have executed.

    // Poll the server waiting for a Succeeded.
    let poll = new Promise((resolve, reject) => {
      let pollOnce = (name, ns, i) => {
        k.readNamespacedPod(name, ns).then(response => {
          let pod = response.body
          if (pod.status == undefined) {
            console.log("Pod not yet scheduled")
            return
          }
          let phase = pod.status.phase
          if (phase == "Succeeded") {
            clearTimers()
            let result = new K8sResult(phase)
            resolve(result)
          } else if (phase == "Failed") {
            clearTimers()
            reject("Pod " + name + " failed to run to completion")
          }
          console.log(pod.metadata.namespace + "/" + pod.metadata.name + " phase " + pod.status.phase)
          // In all other cases we fall through and let the fn be run again.
        }).catch(reason => {
          console.log("failed pod lookup")
          clearTimers()
          reject(reason)
        })
      }
      let interval = setInterval(() => {
        if (cancel) {
          clearInterval(interval)
          clearTimeout(waiter)
          return
        }
        pollOnce(name, ns, interval)
      }, 2000)
      let clearTimers = () => {
        clearInterval(interval)
        clearTimeout(waiter)
      }
    })


    // This will fail if the timelimit is reached.
    let timer = new Promise((solve,ject) => {
      waiter = setTimeout(() => {
        cancel = true
        ject("time limit exceeded")
      }, timeout)
    })

    return Promise.race([poll, timer])
  }
  /**
   * cachePVC builds a persistent volume claim for storing a job's cache.
   *
   * A cache PVC persists between builds. So this is addressable as a Job on a Project.
   */
  protected cachePVC(): kubernetes.V1PersistentVolumeClaim {
    let s = new kubernetes.V1PersistentVolumeClaim()
    s.metadata = new kubernetes.V1ObjectMeta()
    s.metadata.name = this.cacheName()
    s.metadata.labels = {
      "heritage": "brigade",
      "component": "jobCache",
      "job": this.job.name,
      "project": this.project.id
    }

    s.spec = new kubernetes.V1PersistentVolumeClaimSpec()
    s.spec.accessModes = ["ReadWriteMany"]

    let res = new kubernetes.V1ResourceRequirements()
    res.requests = { storage: this.job.cache.size }
    s.spec.resources = res

    return s
  }

}

function sidecarSpec(e: BrigadeEvent, local: string, image: string, project: Project, secName: string): kubernetes.V1Container {
  var imageTag = image
  let repoURL = project.repo.cloneURL

  if (!imageTag) {
    imageTag = "deis/git-sidecar:latest"
  }

  let spec = new kubernetes.V1Container()
  spec.name = "vcs-sidecar",
  spec.env = [
    envVar("VCS_REPO", repoURL),
    envVar("VCS_LOCAL_PATH", local),
    envVar("VCS_REVISION", e.commit),
    envVar("VCS_AUTH_TOKEN", project.repo.token)
  ]
  spec.image = imageTag
  spec.imagePullPolicy = "IfNotPresent",
  spec.volumeMounts = [
    volumeMount("vcs-sidecar", local)
  ]

  if (project.repo.sshKey) {
    spec.env.push({
      name: "BRIGADE_REPO_KEY",
      valueFrom: {
        secretKeyRef: {
          key: "brigadeSSHKey",
          name: secName
        }
      }
    } as kubernetes.V1EnvVar)
  }

  return spec
}

function newRunnerPod(podname: string, brigadeImage: string, imageForcePull: boolean): kubernetes.V1Pod {
  let pod = new kubernetes.V1Pod()
  pod.metadata = new kubernetes.V1ObjectMeta()
  pod.metadata.name = podname
  pod.metadata.labels = {
    "heritage": "brigade",
    "component": "job",
  }

  let c1 = new kubernetes.V1Container()
  c1.name = "brigaderun"
  c1.image = brigadeImage
  c1.command = ["/bin/sh", "/hook/main.sh"]
  c1.imagePullPolicy = imageForcePull ? "Always" : "IfNotPresent"
  c1.securityContext = new kubernetes.V1SecurityContext()

  pod.spec = new kubernetes.V1PodSpec()
  pod.spec.containers = [c1]
  pod.spec.restartPolicy = "Never"
  pod.spec.serviceAccount = serviceAccount
  pod.spec.serviceAccountName = serviceAccount
  return pod
}


function newSecret(name: string): kubernetes.V1Secret {
  let s = new kubernetes.V1Secret()
  s.metadata = new kubernetes.V1ObjectMeta()
  s.metadata.name = name
  s.metadata.labels = {
    "heritage": "brigade",
    "component": "job",
  }
  s.data = {} //{"main.sh": b64enc("echo hello && echo goodbye")}

  return s
}

function envVar(key: string, value: string): kubernetes.V1EnvVar {
  let e = new kubernetes.V1EnvVar()
  e.name = key
  e.value = value
  return e
}

function volumeMount(name: string, mountPath: string): kubernetes.V1VolumeMount {
  let v = new kubernetes.V1VolumeMount()
  v.name = name
  v.mountPath = mountPath
  return v
}


export function b64enc(original: string): string {
  return Buffer.from(original).toString('base64')
}

export function b64dec(encoded: string): string {
  return Buffer.from(encoded, "base64").toString("utf8")
}

function generateScript(job: jobs.Job): string | null {
  if (job.tasks.length == 0) {
    return null
  }
  let newCmd = "#!" + job.shell + "\n\n"

  // if shells that support the `set` command are selected, let's add some sane defaults
  if (job.shell == "/bin/sh" || job.shell == "/bin/bash") {
    newCmd += "set -e\n\n"
  }

  // Join the tasks to make a new command:
  if (job.tasks) {
    newCmd += job.tasks.join("\n")
  }
  return newCmd
}

/**
 * secretToProject transforms a properly formatted Secret into a Project.
 *
 * This is exported for testability, and is not considered part of the stable API.
 */
export function secretToProject(ns: string, secret: kubernetes.V1Secret): Project {
  let p: Project = {
    id: secret.metadata.name,
    name: b64dec(secret.data.repository),
    kubernetes: {
      namespace: secret.metadata.namespace || ns,
      vcsSidecar: ""
    },
    repo: {
      name: secret.metadata.annotations["projectName"],
      cloneURL: null,
    },
    secrets: {},
    allowPrivilegedJobs: true,
    allowHostMounts: false
  }
  if (secret.data.vcsSidecar) {
    p.kubernetes.vcsSidecar = b64dec(secret.data.vcsSidecar)
  }
  if (secret.data.cloneURL) {
    p.repo.cloneURL = b64dec(secret.data.cloneURL)
  }
  if (secret.data.secrets) {
    p.secrets = JSON.parse(b64dec(secret.data.secrets))
  }
  if (secret.data.allowPrivilegedJobs) {
    p.allowPrivilegedJobs = (b64dec(secret.data.allowPrivilegedJobs) == 'true')
  }
  if (secret.data.allowHostMounts) {
    p.allowHostMounts = (b64dec(secret.data.allowHostMounts) == 'true')
  }
  if (secret.data.sshKey) {
    p.repo.sshKey = b64dec(secret.data.sshKey)
  }
  if (secret.data["github.token"]) {
    p.repo.token = b64dec(secret.data["github.token"])
  }
  return p
}
