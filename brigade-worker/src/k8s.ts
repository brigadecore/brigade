/**
 * k8s contains the Kubernetes implementation of Brigade.
 */

/** */

import * as kubernetes from "@kubernetes/client-node";
import * as jobs from "@brigadecore/brigadier/out/job";
import { LogLevel, ContextLogger } from "@brigadecore/brigadier/out/logger";
import { BrigadeEvent, Project } from "@brigadecore/brigadier/out/events";
import * as fs from "fs";
import * as path from "path";
import * as request from "request";
import * as byline_1 from "byline";

// The internals for running tasks. This must be loaded before any of the
// objects that use run().
//
// All Kubernetes API calls should be localized here. Other modules should not
// call 'kubernetes' directly.

// expiresInMSec is the number of milliseconds until pod expiration
// After this point, the pod can be garbage collected (a feature not yet implemented)
const expiresInMSec = 1000 * 60 * 60 * 24 * 30;

const defaultClient = kubernetes.Config.defaultClient();
const retry = (fn, args, delay, times) => {
  // exponential back-off retry if status is in the 500s
  return fn.apply(defaultClient, args).catch(err => {
    if (
      times > 0 &&
      err.response &&
      500 <= err.response.statusCode &&
      err.response.statusCode < 600
    ) {
      return new Promise(resolve => {
        setTimeout(() => {
          resolve(retry(fn, args, delay * 2, times - 1));
        }, delay);
      });
    }
    return Promise.reject(err);
  });
};
const wrapClient = fns => {
  // wrap client methods with retry logic
  for (let fn of fns) {
    let originalFn = defaultClient[fn.name];
    defaultClient[fn.name] = function () {
      return retry(originalFn, arguments, 4000, 5);
    };
  }
};
wrapClient([
  defaultClient.createNamespacedPersistentVolumeClaim,
  defaultClient.deleteNamespacedPersistentVolumeClaim,
  defaultClient.readNamespacedSecret,
  defaultClient.readNamespacedPodLog,
  defaultClient.createNamespacedSecret,
  defaultClient.createNamespacedPod,
  defaultClient.readNamespacedPersistentVolumeClaim,
  defaultClient.deleteNamespacedPod
]);

const getKubeConfig = (): kubernetes.KubeConfig => {
  const kc = new kubernetes.KubeConfig();
  const config =
    process.env.KUBECONFIG || path.join(process.env.HOME, ".kube", "config");
  if (fs.existsSync(config)) {
    kc.loadFromFile(config);
  } else {
    kc.loadFromCluster();
  }
  return kc;
};
const kc = getKubeConfig();

/**
 * options is the set of configuration options for the library.
 *
 * The k8s library provides a backend for the brigade.js objects. But it needs
 * some configuration that is to be passed directly to the library, not via the
 * brigade.js. To allow for this plus overrides (e.g. by Project or Job objects),
 * we maintain a top-level singleton object that holds configuration.
 *
 * It is initially populated with defaults. The defaults can be overridden first
 * by the app (app.ts), then by the project (where allowed). Certain jobs may be
 * allowed to override (or ignore) 'options', though they should never modify
 * it.
 */
export var options: KubernetesOptions = {
  serviceAccount: "brigade-worker",
  mountPath: "/src",
  defaultBuildStorageClass: "",
  defaultCacheStorageClass: ""
};

/**
 * KubernetesOptions exposes options for Kubernetes configuration.
 */
export class KubernetesOptions {
  serviceAccount: string;
  mountPath: string;
  defaultBuildStorageClass: string;
  defaultCacheStorageClass: string;
}

class K8sResult implements jobs.Result {
  data: string;
  constructor(msg: string) {
    this.data = msg;
  }
  toString(): string {
    return this.data;
  }
}

/**
 * BuildStorage manages per-build storage for a build.
 *
 * BuildStorage implements the app.BuildStorage interface.
 *
 * Storage is implemented as a PVC. The PVC backing it MUST be ReadWriteMany.
 */
export class BuildStorage {
  proj: Project;
  name: string;
  build: string;
  logger: ContextLogger = new ContextLogger("k8s");
  options: KubernetesOptions = Object.assign({}, options);

  /**
   * create initializes a new PVC for storing data.
   */
  public create(
    e: BrigadeEvent,
    project: Project,
    size: string
  ): Promise<string> {
    this.proj = project;
    this.name = e.workerID.toLowerCase();
    this.build = e.buildID;
    this.logger.logLevel = e.logLevel;

    let pvc = this.buildPVC(size);
    this.logger.log(`Creating PVC named ${this.name}`);
    return Promise.resolve<string>(
      defaultClient
        .createNamespacedPersistentVolumeClaim(
          this.proj.kubernetes.namespace,
          pvc
        )
        .then(() => {
          return this.name;
        })
    );
  }
  /**
   * destroy deletes the PVC.
   */
  public destroy(): Promise<boolean> {
    if (!this.proj && !this.name) {
      this.logger.log("Build storage not exists");
      return Promise.resolve(false);
    }

    this.logger.log(`Destroying PVC named ${this.name}`);
    let opts = new kubernetes.V1DeleteOptions();
    return Promise.resolve<boolean>(
      defaultClient
        .deleteNamespacedPersistentVolumeClaim(
          this.name,
          this.proj.kubernetes.namespace,
          "true",
          opts
        )
        .then(() => {
          return true;
        })
    );
  }
  /**
   * Get a PVC for a volume that lives for the duration of a build.
   */
  protected buildPVC(size: string): kubernetes.V1PersistentVolumeClaim {
    let s = new kubernetes.V1PersistentVolumeClaim();
    s.metadata = new kubernetes.V1ObjectMeta();
    s.metadata.name = this.name;
    s.metadata.labels = {
      heritage: "brigade",
      component: "buildStorage",
      project: this.proj.id,
      worker: this.name,
      build: this.build
    };

    s.spec = new kubernetes.V1PersistentVolumeClaimSpec();
    s.spec.accessModes = ["ReadWriteMany"];

    let res = new kubernetes.V1ResourceRequirements();
    res.requests = { storage: size };
    s.spec.resources = res;
    if (this.proj.kubernetes.buildStorageClass.length > 0) {
      s.spec.storageClassName = this.proj.kubernetes.buildStorageClass;
    } else if (
      this.options.defaultBuildStorageClass &&
      this.options.defaultBuildStorageClass.length > 0
    ) {
      s.spec.storageClassName = this.options.defaultBuildStorageClass;
    }
    return s;
  }
}

/**
 * loadProject takes a Secret name and namespace and loads the Project
 * from the secret.
 */
export function loadProject(name: string, ns: string): Promise<Project> {
  return Promise.resolve<Project>(
    defaultClient
      .readNamespacedSecret(name, ns)
      .catch(reason => {
        const msg = reason.body ? reason.body.message : reason;
        return Promise.reject(new Error(`Project not found: ${msg}`));
      })
      .then(result => {
        return secretToProject(ns, result.body);
      })
  );
}

/**
 * JobRunner provides a Kubernetes implementation of the JobRunner interface.
 */
export class JobRunner implements jobs.JobRunner {
  name: string;
  secret: kubernetes.V1Secret;
  runner: kubernetes.V1Pod;
  pvc: kubernetes.V1PersistentVolumeClaim;
  project: Project;
  event: BrigadeEvent;
  job: jobs.Job;
  client: kubernetes.CoreV1Api;
  options: KubernetesOptions;
  serviceAccount: string;
  logger: ContextLogger;
  pod: kubernetes.V1Pod;
  cancel: boolean;
  reconnect: boolean;

  constructor() { }

  /**
   * init takes a generic so we can run this against mocks as well as against the real Job type.
   * @param job The Job object
   * @param e  The event that was fired
   * @param project  The project in which this job runs
   * @param allowSecretKeyRef  Allow secretKeyRef in the job's environment
   */
  public init<T extends jobs.Job>(job: T, e: BrigadeEvent, project: Project, allowSecretKeyRef: boolean = true) {
    this.options = Object.assign({}, options);
    this.event = e;
    this.logger = new ContextLogger("k8s", e.logLevel);
    this.job = job;
    this.project = project;
    this.client = defaultClient;
    this.serviceAccount = job.serviceAccount || this.options.serviceAccount;
    this.pod = undefined;
    this.cancel = false;
    this.reconnect = false;

    // $JOB-$BUILD
    this.name = `${job.name}-${this.event.buildID}`;
    let commit = e.revision.commit || "master";
    let secName = this.name;
    let runnerName = this.name;

    this.secret = newSecret(secName);
    this.runner = newRunnerPod(
      runnerName,
      job.image,
      job.imageForcePull,
      this.serviceAccount,
      job.resourceRequests,
      job.resourceLimits,
      job.annotations,
      job.shell
    );

    // Experimenting with setting a deadline field after which something
    // can clean up existing builds.
    let expiresAt = Date.now() + expiresInMSec;

    this.runner.metadata.labels.jobname = job.name;
    this.runner.metadata.labels.project = project.id;
    this.runner.metadata.labels.worker = e.workerID;
    this.runner.metadata.labels.build = e.buildID;

    this.secret.metadata.labels.jobname = job.name;
    this.secret.metadata.labels.project = project.id;
    this.secret.metadata.labels.expires = String(expiresAt);
    this.secret.metadata.labels.worker = e.workerID;
    this.secret.metadata.labels.build = e.buildID;

    let envVars: kubernetes.V1EnvVar[] = [];
    for (let key in job.env) {
      let val = job.env[key];

      if (typeof val === "string") {
        // For environmental variables that are submitted as strings,
        // add to the job's secret and add a reference.

        this.secret.data[key] = b64enc(val);
        // Add reference to pod
        envVars.push({
          name: key,
          valueFrom: {
            secretKeyRef: {
              name: secName,
              key: key
            }
          }
        } as kubernetes.V1EnvVar);
      } else if (val === null) {
        envVars.push({
          name: key,
          value: ""
        } as kubernetes.V1EnvVar);
      } else {
        // For environmental variables that are directly references,
        // add the reference to the env var list.

        if (val.secretKeyRef != null && !allowSecretKeyRef) {
          // allowSecretKeyRef is not to true so disallow setting secrets in the environment
          this.logger.warn(`Using secretKeyRef is not allowed in this project, not setting environment variable ${key}`);
          continue
        }

        envVars.push({
          name: key,
          valueFrom: val
        } as kubernetes.V1EnvVar);
      }
    }

    this.runner.spec.containers[0].env = envVars;

    let mountPath = job.mountPath || this.options.mountPath;

    // Add secret volume
    this.runner.spec.volumes = [
      { name: secName, secret: { secretName: secName } } as kubernetes.V1Volume
    ];
    this.runner.spec.containers[0].volumeMounts = [
      { name: secName, mountPath: "/hook" } as kubernetes.V1VolumeMount
    ];

    this.runner.spec.initContainers = [];
    if (job.useSource && project.repo.cloneURL && project.kubernetes.vcsSidecar) {
      // Add the sidecar.
      let sidecar = sidecarSpec(
        e,
        "/src",
        project.kubernetes.vcsSidecar,
        project
      );
      this.runner.spec.initContainers = [sidecar];

      // Add volume/volume mounts
      this.runner.spec.volumes.push(
        { name: "vcs-sidecar", emptyDir: {} } as kubernetes.V1Volume
      );
      this.runner.spec.containers[0].volumeMounts.push(
        { name: "vcs-sidecar", mountPath: mountPath } as kubernetes.V1VolumeMount
      );
    }

    if (job.imagePullSecrets) {
      this.runner.spec.imagePullSecrets = [];
      for (let secret of job.imagePullSecrets) {
        this.runner.spec.imagePullSecrets.push({ name: secret });
      }
    }

    // If host os is set, specify it.
    if (job.host.os) {
      this.runner.spec.nodeSelector = {
        "beta.kubernetes.io/os": job.host.os
      };
    }
    if (job.host.name) {
      this.runner.spec.nodeName = job.host.name;
    }
    if (job.host.nodeSelector && job.host.nodeSelector.size > 0) {
      if (!this.runner.spec.nodeSelector) {
        this.runner.spec.nodeSelector = {};
      }
      for (const k of job.host.nodeSelector.keys()) {
        this.runner.spec.nodeSelector[k] = job.host.nodeSelector.get(k);
      }
    }

    // If the job requests a cache, set up the cache.
    if (job.cache.enabled) {
      this.pvc = this.cachePVC();

      // Now add volume mount to pod:
      let mountName = this.cacheName();
      this.runner.spec.volumes.push({
        name: mountName,
        persistentVolumeClaim: { claimName: mountName }
      } as kubernetes.V1Volume);
      let mnt = volumeMount(mountName, job.cache.path);
      this.runner.spec.containers[0].volumeMounts.push(mnt);
    }

    // If the job needs build-wide storage, enable it.
    if (job.storage.enabled) {
      const vname = "build-storage";
      this.runner.spec.volumes.push({
        name: vname,
        persistentVolumeClaim: { claimName: e.workerID.toLowerCase() }
      } as kubernetes.V1Volume);
      let mnt = volumeMount(vname, job.storage.path);
      this.runner.spec.containers[0].volumeMounts.push(mnt);
    }

    // If the job needs access to a docker daemon, mount in the host's docker socket
    if (job.docker.enabled && project.allowHostMounts) {
      var dockerVol = new kubernetes.V1Volume();
      var dockerMount = new kubernetes.V1VolumeMount();
      var hostPath = new kubernetes.V1HostPathVolumeSource();
      hostPath.path = jobs.dockerSocketMountPath;
      dockerVol.name = jobs.dockerSocketMountName;
      dockerVol.hostPath = hostPath;
      dockerMount.name = jobs.dockerSocketMountName;
      dockerMount.mountPath = jobs.dockerSocketMountPath;
      this.runner.spec.volumes.push(dockerVol);
      for (let i = 0; i < this.runner.spec.containers.length; i++) {
        this.runner.spec.containers[i].volumeMounts.push(dockerMount);
      }
    }

    if (job.args.length > 0) {
      this.runner.spec.containers[0].args = job.args;
    }

    let newCmd = generateScript(job);
    if (!newCmd) {
      this.runner.spec.containers[0].command = null;
    } else {
      this.secret.data["main.sh"] = b64enc(newCmd);
    }

    // If the job askes for privileged mode and the project allows this, enable it.
    if (job.privileged && project.allowPrivilegedJobs) {
      for (let i = 0; i < this.runner.spec.containers.length; i++) {
        this.runner.spec.containers[i].securityContext.privileged = true;
      }
    }
    return this;
  }

  /**
   * cacheName returns the name of this job's cache PVC.
   */
  protected cacheName(): string {
    // The Kubernetes rules on pvc names are stupid^b^b^b^b strict. Name must
    // be DNS-like, and less than 64 chars. This rules out using project ID,
    // project name, etc. For now, we use project name with slashes replaced,
    // appended to job name.
    return `${this.project.name.replace(/[.\/]/g, "-")}-${
      this.job.name
      }`.toLowerCase();
  }

  public logs(): Promise<string> {
    let podName = this.name;
    let k = this.client;
    let ns = this.project.kubernetes.namespace;
    return Promise.resolve<string>(
      k.readNamespacedPodLog(podName, ns).then(result => {
        return result.body;
      })
    );
  }

  /**
   * run starts a job and then waits until it is running.
   *
   * The Promise it returns will return when the pod is either marked
   * Success (resolve) or Failure (reject)
   */
  public run(): Promise<jobs.Result> {
    return this.start()
      .then(r => r.wait())
      .then(r => {
        return this.logs();
      })
      .then(response => {
        return new K8sResult(response);
      });
  }

  /** start begins a job, and returns once it is scheduled to run.*/
  public start(): Promise<jobs.JobRunner> {
    // Now we have pod and a secret defined. Time to create them.

    let ns = this.project.kubernetes.namespace;
    let k = this.client;
    let pvcPromise = this.checkOrCreateCache();

    return new Promise((resolve, reject) => {
      pvcPromise
        .then(() => {
          this.logger.log("Creating secret " + this.secret.metadata.name);
          return k.createNamespacedSecret(ns, this.secret);
        })
        .then(result => {
          this.logger.log("Creating pod " + this.runner.metadata.name);
          // Once namespace creation has been accepted, we create the pod.
          return k.createNamespacedPod(ns, this.runner);
        })
        .then(result => {
          resolve(this);
        })
        .catch(reason => {
          reject(new Error(reason.body.message));
        });
    });
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
      let ns = this.project.kubernetes.namespace;
      let k = this.client;
      if (!this.pvc) {
        resolve("no cache requested");
      } else {
        let cname = this.cacheName();
        this.logger.log(`looking up ${ns}/${cname}`);
        k.readNamespacedPersistentVolumeClaim(cname, ns)
          .then(result => {
            resolve("re-using existing cache");
          })
          .catch(result => {
            // TODO: check if cache exists.
            this.logger.log(`Creating Job Cache PVC ${cname}`);
            return k
              .createNamespacedPersistentVolumeClaim(ns, this.pvc)
              .then(result => {
                this.logger.log("created cache");
                resolve("created job cache");
              });
          })
          .catch(err => {
            reject(new Error(err.body.message));
          });
      }
    });
  }

  /**
   * update pod info on event using watch
   */
  private startUpdatingPod(): request.Request {
    const url = `${kc.getCurrentCluster().server}/api/v1/namespaces/${
      this.project.kubernetes.namespace
      }/pods`;
    const requestOptions = {
      qs: {
        watch: true,
        timeoutSeconds: 200,
        labelSelector: `build=${this.event.buildID},jobname=${this.job.name}`
      },
      method: "GET",
      uri: url,
      useQuerystring: true,
      json: true
    };
    kc.applyToRequest(requestOptions);
    const stream = new byline_1.LineStream();
    stream.on("data", data => {
      let obj = null;
      try {
        if (data instanceof Buffer) {
          obj = JSON.parse(data.toString());
        } else {
          obj = JSON.parse(data);
        }
      } catch (e) { } //let it stay connected.
      if (obj && obj.object) {
        this.pod = obj.object as kubernetes.V1Pod;
      }
    });
    const req = request(requestOptions, (error, response, body) => {
      if (error) {
        this.logger.error(error.body.message);
        this.reconnect = true; //reconnect unless aborted
      }
    });
    req.pipe(stream);
    req.on("end", () => {
      this.reconnect = true;
    }); //stay connected on transient faults
    return req;
  }

  /** wait listens for the running job to complete.*/
  public wait(): Promise<jobs.Result> {
    // Should probably protect against the case where start() was not called
    let k = this.client;
    let timeout = this.job.timeout || 60000;
    let name = this.name;
    let ns = this.project.kubernetes.namespace;
    let podUpdater: request.Request = undefined;

    // This is a handle to clear the setTimeout when the promise is fulfilled.
    let waiter;
    // Handle to abort the request on completion and only to ensure that we hook the 'follow logs' events only once
    let followLogsRequest: request.Request = null;

    this.logger.log(`Timeout set at ${timeout} milliseconds`);

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
        if (!podUpdater) {
          podUpdater = this.startUpdatingPod();
        } else if (!this.cancel && this.reconnect) {
          //if not intentionally cancelled, reconnect
          this.reconnect = false;
          try {
            podUpdater.abort();
          } catch (e) {
            this.logger.log(e);
          }
          podUpdater = this.startUpdatingPod();
        }
        if (!this.pod || this.pod.status == undefined) {
          this.logger.log("Pod not yet scheduled");
          return;
        }

        let phase = this.pod.status.phase;
        if (phase == "Succeeded") {
          clearTimers();
          let result = new K8sResult(phase);
          resolve(result);
        }
        // make sure Pod is running before we start following its logs
        else if (phase == "Running") {
          // do that only if we haven't hooked up the follow request before
          if (followLogsRequest == null && this.job.streamLogs) {
            followLogsRequest = followLogs(this.pod.metadata.namespace, this.pod.metadata.name);
          }
        } else if (phase == "Failed") {
          clearTimers();
          reject(new Error(`Pod ${name} failed to run to completion`));
        } else if (phase == "Pending") {
          // Trap image pull errors and consider them fatal.
          let cs = this.pod.status.containerStatuses;
          if (
            cs &&
            cs.length > 0 &&
            cs[0].state.waiting &&
            cs[0].state.waiting.reason == "ErrImagePull"
          ) {
            k.deleteNamespacedPod(
              name,
              ns,
              "true",
              new kubernetes.V1DeleteOptions()
            ).catch(e => this.logger.error(e.body.message));
            clearTimers();
            reject(new Error(cs[0].state.waiting.message));
          }
        }
        if (!this.job.streamLogs || (this.job.streamLogs && this.pod.status.phase != "Running")) {
          // don't display "Running" when we're asked to display job Pod logs
          this.logger.log(`${this.pod.metadata.namespace}/${this.pod.metadata.name} phase ${this.pod.status.phase}`);
        }
        // In all other cases we fall through and let the fn be run again.
      };
      let interval = setInterval(() => {
        if (this.cancel) {
          podUpdater.abort();
          clearInterval(interval);
          clearTimeout(waiter);
          return;
        }
        pollOnce(name, ns, interval);
      }, 2000);
      let clearTimers = () => {
        podUpdater.abort();
        clearInterval(interval);
        clearTimeout(waiter);
        if (followLogsRequest != null) {
          followLogsRequest.abort();
        }
      };

      // follows logs for the specified namespace/Pod combination
      let followLogs = (namespace: string, podName: string): request.Request => {
        const url = `${kc.getCurrentCluster().server}/api/v1/namespaces/${namespace}/pods/${podName}/log`;
        //https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#pod-v1-core
        const requestOptions = {
          qs: {
            follow: true,
            timeoutSeconds: 200,
          },
          method: "GET",
          uri: url,
          useQuerystring: true
        };
        kc.applyToRequest(requestOptions);
        const stream = new byline_1.LineStream();
        stream.on("data", data => {
          let logs = null;
          try {
            if (data instanceof Buffer) {
              logs = data.toString();
            } else {
              logs = data;
            }
            this.logger.log(
              `${this.pod.metadata.namespace}/${this.pod.metadata.name} logs ${logs}`
            );
          } catch (e) { } //let it stay connected.
        });
        const req = request(requestOptions, (error, response, body) => {
          if (error) {
            this.logger.error(error.body.message);
            this.reconnect = true; //reconnect unless aborted
          }
        });
        req.pipe(stream);
        return req;
      }

    });

    // This will fail if the timelimit is reached.
    let timer = new Promise((solve, reject) => {
      waiter = setTimeout(() => {
        this.cancel = true;
        reject("time limit exceeded");
      }, timeout);
    });

    return Promise.race([poll, timer]);
  }
  /**
   * cachePVC builds a persistent volume claim for storing a job's cache.
   *
   * A cache PVC persists between builds. So this is addressable as a Job on a Project.
   */
  protected cachePVC(): kubernetes.V1PersistentVolumeClaim {
    let s = new kubernetes.V1PersistentVolumeClaim();
    s.metadata = new kubernetes.V1ObjectMeta();
    s.metadata.name = this.cacheName();
    s.metadata.labels = {
      heritage: "brigade",
      component: "jobCache",
      job: this.job.name,
      project: this.project.id
    };

    s.spec = new kubernetes.V1PersistentVolumeClaimSpec();
    s.spec.accessModes = ["ReadWriteMany"];
    if (
      this.project.kubernetes.cacheStorageClass &&
      this.project.kubernetes.cacheStorageClass.length > 0
    ) {
      s.spec.storageClassName = this.project.kubernetes.cacheStorageClass;
    } else if (
      this.options.defaultCacheStorageClass &&
      this.options.defaultCacheStorageClass.length > 0
    ) {
      s.spec.storageClassName = this.options.defaultCacheStorageClass;
    }
    let res = new kubernetes.V1ResourceRequirements();
    res.requests = { storage: this.job.cache.size };
    s.spec.resources = res;

    return s;
  }
}

function sidecarSpec(
  e: BrigadeEvent,
  local: string,
  image: string,
  project: Project
): kubernetes.V1Container {
  var imageTag = image;
  let initGitSubmodules = project.repo.initGitSubmodules;

  if (!imageTag) {
    imageTag = "brigadecore/git-sidecar:latest";
  }

  let spec = new kubernetes.V1Container();
  (spec.name = "vcs-sidecar"),
    (spec.env = [
      envVar("CI", "true"),
      envVar("BRIGADE_BUILD_ID", e.buildID),
      envVar("BRIGADE_COMMIT_ID", e.revision.commit),
      envVar("BRIGADE_COMMIT_REF", e.revision.ref),
      envVar("BRIGADE_EVENT_PROVIDER", e.provider),
      envVar("BRIGADE_EVENT_TYPE", e.type),
      envVar("BRIGADE_PROJECT_ID", project.id),
      envVar("BRIGADE_REMOTE_URL", project.repo.cloneURL),
      envVar("BRIGADE_WORKSPACE", local),
      envVar("BRIGADE_PROJECT_NAMESPACE", project.kubernetes.namespace),
      envVar("BRIGADE_SUBMODULES", initGitSubmodules.toString()),
      envVar("BRIGADE_LOG_LEVEL", LogLevel[e.logLevel])
    ]);
  spec.image = imageTag;
  (spec.imagePullPolicy = "IfNotPresent"),
    (spec.volumeMounts = [volumeMount("vcs-sidecar", local)]);

  if (project.repo.sshKey) {
    spec.env.push({
      name: "BRIGADE_REPO_KEY",
      valueFrom: {
        secretKeyRef: {
          key: "sshKey",
          name: project.id
        }
      }
    } as kubernetes.V1EnvVar);
  }

  if (project.repo.token) {
    spec.env.push({
      name: "BRIGADE_REPO_AUTH_TOKEN",
      valueFrom: {
        secretKeyRef: {
          key: "github.token",
          name: project.id
        }
      }
    } as kubernetes.V1EnvVar);
  }

  spec.resources = new kubernetes.V1ResourceRequirements();
  spec.resources.limits = {};
  spec.resources.requests = {};
  if (project.kubernetes.vcsSidecarResourcesLimitsCPU) {
    spec.resources.limits["cpu"] =
      project.kubernetes.vcsSidecarResourcesLimitsCPU;
  }
  if (project.kubernetes.vcsSidecarResourcesLimitsMemory) {
    spec.resources.limits["memory"] =
      project.kubernetes.vcsSidecarResourcesLimitsMemory;
  }
  if (project.kubernetes.vcsSidecarResourcesRequestsCPU) {
    spec.resources.requests["cpu"] =
      project.kubernetes.vcsSidecarResourcesRequestsCPU;
  }
  if (project.kubernetes.vcsSidecarResourcesRequestsMemory) {
    spec.resources.requests["memory"] =
      project.kubernetes.vcsSidecarResourcesRequestsMemory;
  }

  return spec;
}

function newRunnerPod(
  podname: string,
  brigadeImage: string,
  imageForcePull: boolean,
  serviceAccount: string,
  resourceRequests: jobs.JobResourceRequest,
  resourceLimits: jobs.JobResourceLimit,
  jobAnnotations: { [key: string]: string },
  jobShell: string
): kubernetes.V1Pod {
  let pod = new kubernetes.V1Pod();
  pod.metadata = new kubernetes.V1ObjectMeta();
  pod.metadata.name = podname;
  pod.metadata.labels = {
    heritage: "brigade",
    component: "job"
  };
  pod.metadata.annotations = jobAnnotations;

  let c1 = new kubernetes.V1Container();
  c1.name = "brigaderun";
  c1.image = brigadeImage;

  if (jobShell == "") {
    jobShell = "/bin/sh";
  }
  c1.command = [jobShell, "/hook/main.sh"];

  c1.imagePullPolicy = imageForcePull ? "Always" : "IfNotPresent";
  c1.securityContext = new kubernetes.V1SecurityContext();

  // Setup pod container resources (requests and limits).
  let resourceRequirements = new kubernetes.V1ResourceRequirements();
  if (resourceRequests) {
    resourceRequirements.requests = {
      cpu: resourceRequests.cpu,
      memory: resourceRequests.memory
    };
  }
  if (resourceLimits) {
    resourceRequirements.limits = {
      cpu: resourceLimits.cpu,
      memory: resourceLimits.memory
    };
  }

  c1.resources = resourceRequirements;
  pod.spec = new kubernetes.V1PodSpec();
  pod.spec.containers = [c1];
  pod.spec.restartPolicy = "Never";
  pod.spec.serviceAccount = serviceAccount;
  pod.spec.serviceAccountName = serviceAccount;
  return pod;
}

function newSecret(name: string): kubernetes.V1Secret {
  let s = new kubernetes.V1Secret();
  s.type = "brigade.sh/job";
  s.metadata = new kubernetes.V1ObjectMeta();
  s.metadata.name = name;
  s.metadata.labels = {
    heritage: "brigade",
    component: "job"
  };
  s.data = {}; //{"main.sh": b64enc("echo hello && echo goodbye")}

  return s;
}

function envVar(key: string, value: string): kubernetes.V1EnvVar {
  let e = new kubernetes.V1EnvVar();
  e.name = key;
  e.value = value;
  return e;
}

function volumeMount(
  name: string,
  mountPath: string
): kubernetes.V1VolumeMount {
  let v = new kubernetes.V1VolumeMount();
  v.name = name;
  v.mountPath = mountPath;
  return v;
}

export function b64enc(original: string): string {
  return Buffer.from(original).toString("base64");
}

export function b64dec(encoded: string): string {
  return Buffer.from(encoded, "base64").toString("utf8");
}

function generateScript(job: jobs.Job): string | null {
  if (job.tasks.length == 0) {
    return null;
  }
  let newCmd = "#!" + job.shell + "\n\n";

  // if shells that support the `set` command are selected, let's add some sane defaults
  if (job.shell == "/bin/sh" || job.shell == "/bin/bash") {
    newCmd += "set -e\n\n";
  }

  // Join the tasks to make a new command:
  if (job.tasks) {
    newCmd += job.tasks.join("\n");
  }
  return newCmd;
}

/**
 * secretToProject transforms a properly formatted Secret into a Project.
 *
 * This is exported for testability, and is not considered part of the stable API.
 */
export function secretToProject(
  ns: string,
  secret: kubernetes.V1Secret
): Project {
  let p: Project = {
    id: secret.metadata.name,
    name: secret.metadata.annotations["projectName"],
    kubernetes: {
      namespace: secret.metadata.namespace || ns,
      buildStorageSize: "50Mi",
      vcsSidecar: "",
      vcsSidecarResourcesLimitsCPU: "",
      vcsSidecarResourcesLimitsMemory: "",
      vcsSidecarResourcesRequestsCPU: "",
      vcsSidecarResourcesRequestsMemory: "",
      cacheStorageClass: "",
      buildStorageClass: ""
    },
    secrets: {},
    allowPrivilegedJobs: true,
    allowHostMounts: false
  };
  if (secret.data.repository != null) {
    // For legacy/backwards-compatibility reasons,
    // we set project name and repo name to the following values,
    // despite the fact that they should logically be swapped.
    p.name = b64dec(secret.data.repository)
    p.repo = {
      name: secret.metadata.annotations["projectName"],
      cloneURL: null,
      initGitSubmodules: false
    }
  }
  if (secret.data.vcsSidecar) {
    p.kubernetes.vcsSidecar = b64dec(secret.data.vcsSidecar);
  }
  if (secret.data["vcsSidecarResources.limits.cpu"]) {
    p.kubernetes.vcsSidecarResourcesLimitsCPU = b64dec(
      secret.data["vcsSidecarResources.limits.cpu"]
    );
  }
  if (secret.data["vcsSidecarResources.limits.memory"]) {
    p.kubernetes.vcsSidecarResourcesLimitsMemory = b64dec(
      secret.data["vcsSidecarResources.limits.memory"]
    );
  }
  if (secret.data["vcsSidecarResources.requests.cpu"]) {
    p.kubernetes.vcsSidecarResourcesRequestsCPU = b64dec(
      secret.data["vcsSidecarResources.requests.cpu"]
    );
  }
  if (secret.data["vcsSidecarResources.requests.memory"]) {
    p.kubernetes.vcsSidecarResourcesRequestsMemory = b64dec(
      secret.data["vcsSidecarResources.requests.memory"]
    );
  }
  if (secret.data.buildStorageSize) {
    p.kubernetes.buildStorageSize = b64dec(secret.data.buildStorageSize);
  }
  if (secret.data.cloneURL) {
    p.repo.cloneURL = b64dec(secret.data.cloneURL);
  }
  if (secret.data.initGitSubmodules) {
    p.repo.initGitSubmodules = b64dec(secret.data.initGitSubmodules) == "true";
  }
  if (secret.data.secrets) {
    p.secrets = JSON.parse(b64dec(secret.data.secrets));
  }
  if (secret.data.allowPrivilegedJobs) {
    p.allowPrivilegedJobs = b64dec(secret.data.allowPrivilegedJobs) == "true";
  }
  if (secret.data.allowHostMounts) {
    p.allowHostMounts = b64dec(secret.data.allowHostMounts) == "true";
  }
  if (secret.data.sshKey) {
    p.repo.sshKey = b64dec(secret.data.sshKey);
  }
  if (secret.data["github.token"]) {
    p.repo.token = b64dec(secret.data["github.token"]);
  }
  if (secret.data["kubernetes.cacheStorageClass"]) {
    p.kubernetes.cacheStorageClass = b64dec(
      secret.data["kubernetes.cacheStorageClass"]
    );
  }
  if (secret.data["kubernetes.buildStorageClass"]) {
    p.kubernetes.buildStorageClass = b64dec(
      secret.data["kubernetes.buildStorageClass"]
    );
  }
  return p;
}
