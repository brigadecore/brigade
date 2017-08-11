import * as kubernetes from '@kubernetes/typescript-node'
import * as jobs from './job'
import {AcidEvent, Project} from './events'

// The internals for running tasks. This must be loaded before any of the
// objects that use run().
//
// All Kubernetes API calls should be localized here. Other modules should not
// call 'kubernetes' directly.

// expiresInMSec is the number of milliseconds until pod expiration
// After this point, the pod can be garbage collected (a feature not yet implemented)
const expiresInMSec = 1000 * 60 * 60 * 24 * 30

const defaultClient = kubernetes.Config.defaultClient()

class K8sResult implements jobs.Result {
  data: string
  constructor(msg: string) { this.data = msg }
  toString(): string { return this.data }
}

export function loadProject(name: string, ns: string): Promise<Project> {
  return defaultClient.readNamespacedSecret(name, ns).then( result => {
    return secretToProject(ns, result.body)
  })
}

export class JobRunner implements jobs.JobRunner {

  name: string
  secret: kubernetes.V1Secret
  runner: kubernetes.V1Pod
  project: Project
  event: AcidEvent
  job: jobs.Job
  client: kubernetes.Core_v1Api

  constructor(job: jobs.Job, e: AcidEvent, project: Project) {
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
    this.runner = newRunnerPod(runnerName, job.image)

    let belongsto = project.repo.name.replace(/\//g, "-")

    // Experimenting with setting a deadline field after which something
    // can clean up existing builds.
    let expiresAt = Date.now() + expiresInMSec

    this.runner.metadata.labels.jobname = job.name
    this.runner.metadata.labels.belongsto = belongsto
    this.runner.metadata.labels.commit = commit
    this.runner.metadata.labels.role = "job"

    this.secret.metadata.labels.jobname = job.name
    this.secret.metadata.labels.belongsto = belongsto
    this.secret.metadata.labels.commit = commit
    this.secret.metadata.labels.expires = String(expiresAt)

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
      this.secret.data.acidSSHKey = b64enc(project.repo.sshKey)
      envVars.push({
        name: "ACID_REPO_KEY",
        valueFrom: {
          secretKeyRef: {
            key: "acidSSHKey",
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

    if (job.useSource) {
      // Add the sidecar.
      let sidecar = sidecarSpec(e, "/src", project.kubernetes.vcsSidecar, project, secName)

      // TODO: convert this to an init container with Kube 1.6
      // runner.spec.initContainers = [sidecar]
      this.runner.metadata.annotations = {
        "pod.beta.kubernetes.io/init-containers": "[" + JSON.stringify(sidecar) + "]"
      }
    }

    let newCmd = generateScript(job)
    if (!newCmd) {
      this.runner.spec.containers[0].command = null
    } else {
      this.secret.data["main.sh"] = b64enc(newCmd)
    }
  }

  // run starts a job and then waits until it is running.
  //
  // The Promise it returns will return when the pod is either marked
  // Success (resolve) or Failure (reject)
  public run(): Promise<jobs.Result> {
    return this.start().then(r => r.wait())
  }

  // start begins a job, and returns once it is scheduled to run.
  public start(): Promise<jobs.JobRunner> {
    // Now we have pod and a secret defined. Time to create them.
    let ns = this.project.kubernetes.namespace
    let k = this.client
    return new Promise((resolve, reject) => {
      console.log("Creating secret " + this.secret.metadata.name)
      k.createNamespacedSecret(ns, this.secret).then((result, newSec) => {
        console.log("Creating pod " + this.runner.metadata.name)
        // Once namespace creation has been accepted, we create the pod.
        k.createNamespacedPod(ns, this.runner).then((result, newPod) => {
          resolve(this)
        }).catch(reason => reject(reason))

      }).catch(reason => reject(reason))
    })
  }

  // wait listens for the running job to complete.
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
}

function sidecarSpec(e: AcidEvent, local: string, image: string, project: Project, secName: string): kubernetes.V1Container {
  var imageTag = image
  let repoURL = project.repo.cloneURL

  if (!imageTag) {
    imageTag = "acid/vcs-sidecar:latest"
  }

  let spec = new kubernetes.V1Container()
  spec.name = "acid-vcs-sidecar",
  spec.env = [
    envVar("VCS_REPO", repoURL),
    envVar("VCS_LOCAL_PATH", local),
    envVar("VCS_REVISION", e.commit)
  ]
  spec.image = imageTag
  spec.command = ["/vcs-sidecar"]
  spec.imagePullPolicy = "IfNotPresent",
  spec.volumeMounts = [
    volumeMount("vcs-sidecar", local)
  ]

  if (project.repo.sshKey) {
    spec.env.push({
      name: "ACID_REPO_KEY",
      valueFrom: {
        secretKeyRef: {
          key: "acidSSHKey",
          name: secName
        }
      }
    } as kubernetes.V1EnvVar)
  }

  return spec
}

function newRunnerPod(podname: string, acidImage: string): kubernetes.V1Pod {
  let pod = new kubernetes.V1Pod()
  pod.metadata = new kubernetes.V1ObjectMeta()
  pod.metadata.name = podname
  pod.metadata.labels = {
    "heritage": "acid",
    "managedBy": "acid"
  }

  let c1 = new kubernetes.V1Container()
  c1.name = "acidrun"
  c1.image = acidImage
  c1.command = ["/bin/sh", "/hook/main.sh"]
  c1.imagePullPolicy = "IfNotPresent"

  pod.spec = new kubernetes.V1PodSpec()
  pod.spec.containers = [c1]
  pod.spec.restartPolicy = "Never"
  return pod
}


function newSecret(name: string): kubernetes.V1Secret {
  let s = new kubernetes.V1Secret()
  s.metadata = new kubernetes.V1ObjectMeta()
  s.metadata.name = name
  s.metadata.labels = {"heritage": "acid"}
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

// secretToProject transforms a properly formatted Secret into a Project.
//
// This is exported for testability, and is not considered part of the stable API.
export function secretToProject(ns: string, secret: kubernetes.V1Secret): Project {
  let p: Project = {
    id: secret.metadata.name,
    name: b64dec(secret.data.repository),
    kubernetes: {
      namespace: secret.metadata.namespace || ns,
      vcsSidecar: b64dec(secret.data.vcsSidecar)
    },
    repo: {
      name: secret.metadata.annotations["projectName"],
      cloneURL: b64dec(secret.data.cloneURL)
    },
    secrets: {}
  }
  if (secret.data.secrets) {
    p.secrets = JSON.parse(b64dec(secret.data.secrets))
  }
  return p
}


