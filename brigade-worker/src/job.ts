/**
 * Package job provides support for jobs.
 *
 * A Job idescribes a particular unit of a build. A Job returns a Result.
 * A JobRunner is an implementation of the runtime logic for a Job.
 */

/** */

import {BrigadeEvent, Project} from "./events"

/**
 * The default shell for the job.
 */
const defaultShell: string = '/bin/sh'
/**
 * defaultTimeout is the default timeout for a job (15 minutes)
 */
const defaultTimeout: number = 1000 * 60 * 15
/**
 * The default image if `Job.image` is not set
 */
const brigadeImage: string = 'debian:jessie-slim'

export const brigadeCachePath = '/mnt/brigade/cache'
export const brigadeStoragePath = '/mnt/brigade/share'
export const dockerSocketMountPath = '/var/run/docker.sock'
export const dockerSocketMountName = 'docker-socket'

/**
 * JobRunner is capable of executing a job within a runtime.
 */
export interface JobRunner {
  // TODO: Should we add the constructor here?
  // Start runs a new job. It returns a JobRunner that can be waited upon.
  start(): Promise<JobRunner>
  // Wait waits unti the job being run has reached a success or failure state.
  wait(): Promise<Result>
}

/**
 * Result is the result of a particular Job.
 *
 * Every Result can be converted to a String with the `toString()` function. The
 * string is human-readable.
 */
export interface Result {
  toString(): string
}

/**
 * Cache controls the job's cache.
 *
 * A cache is a small storage space that is shared between different instances
 * if the same job.
 *
 * Cache is just a plain filesystem, and as such comes with no guarantees about
 * consistency, etc. It should be treated as volatile.
 */
export class JobCache {
  /**
   * If enabled=true, a storage cache will be attached.
   */
  public enabled: boolean = false
  /**
   * size is the amount of storage space assigned to the cache. The default is
   * 5Mi.
   * For sizing information, see https://github.com/kubernetes/community/blob/master/contributors/design-proposals/resources.md
   */
  public size: string = "5Mi"

  // future-proof Cache.path. For now we will hard-code it, but make it so that
  // we can modify in the future.
  private _path: string = brigadeCachePath
  public get path(): string { return this._path }
}

/**
 * JobStorage configures build-wide storage preferences for this job.
 *
 * Changes to this object only impact the job, not the entire build.
 */
export class JobStorage {
  public enabled: boolean = false
  private _path: string = brigadeStoragePath
  public get path(): string { return this._path }
}

/**
 * JobHost expresses expectations about the host a job will run on.
 */
export class JobHost {
  /**
   * os is the name of the OS upon which the job's container must run.
   *
   * This allows users to indicate that the container _must_ run on
   * "windows" or "linux" hosts. It is primarily useful in a "mixed node"
   * environment where the brigade.js will be run on a cluster that has more than
   * one OS
   */
  public os: string

  /**
   * name of the host to run on.
   *
   * If this is set, a job will ask to be run on this named host. Generally, this
   * should be used only if it is necessary to run the job on a particular host.
   * If not set, Brigade will let the scheduler decide, which is strongly preferred.
   *
   * Example usage: If you use a Kubernetes-ACI bridge, you may want to use this
   * to run jobs on the bridge.
   */
  public name: string
}

/**
 * JobDockerMount enables or disables mounting the host's docker socket for a job.
 */
export class JobDockerMount {
  /**
   * enabled configues whether or not the job will mount the host's docker socket.
   */
  public enabled: boolean = false
}

 /**
  * Job represents a single job, which is composed of several closely related sequential tasks.
  * Jobs must have names. Every job also has an associated image, which references
  * the Docker container to be run.
  * */
export abstract class Job {
  /** name of the job*/
  public name: string
  /** shell that will be used by default in this job*/
  public shell: string = defaultShell
  /** tasks is a list of tasks run inside of the shell*/
  public tasks: string[]
  /** env is the environment variables for the job*/
  public env: {[key: string]:string}
  /** image is the container image to be run*/
  public image: string = brigadeImage
  /** imageForcePull defines the container image pull policy: Always if true or IfNotPresent if false */
  public imageForcePull: boolean = false
  /**
   * imagePullSecrets names secrets that contain the credentials for pulling this
   * image or the sidecar image.
   */
  public imagePullSecrets: string[]

  /** Path to mount as the base path for executable code in the container.*/
  public mountPath: string = "/src"

  /** Set the max time to wait for this job to complete.*/
  public timeout: number = defaultTimeout

  /** Fetch the source repo. Default: true*/
  public useSource: boolean = true

  /** If true, the job will be run in privileged mode.
   * This is necessary for Docker engines running inside the Job, for example.
   */
  public privileged: boolean = false

  /**
   * host expresses expectations about the host the job will run on.
   */
  public host: JobHost

  /**
   * cache controls per-Job caching preferences.
   */
  public cache: JobCache
  /**
   * storage controls this job's preferences on the build-wide storage.
   */
  public storage: JobStorage

  /**
   * docker controls the job's preferences on mounting the host's docker daemon.
   */
  public docker: JobDockerMount

  /** _podName is set by the runtime. It is the name of the pod.*/
  protected _podName: string

  /** podName is the generated name of the pod.*/
  get podName(): string {
    return this._podName
  }

  /** Create a new Job
   * name is the name of the job.
   * image is the container image to use
   * tasks is a list of commands to run.
   */
  constructor(name: string, image?: string, tasks?: string[], imageForcePull?: boolean) {
    if (!jobNameIsValid(name)) {
      throw "job name must be letters, numbers, and '-', and must not start or end with '-'"
    }
    this.name = name
    this.image = image
    this.imageForcePull = imageForcePull
    this.tasks = tasks || []
    this.env = {}
    this.cache = new JobCache()
    this.storage = new JobStorage()
    this.docker = new JobDockerMount()
    this.host = new JobHost()
  }

  /** run executes the job and then */
  public abstract run(): Promise<Result>
}

/**
 * jobNameIsValid checks the validity of a job's name.
 */
export function jobNameIsValid(name: string): boolean {
  return /^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])+$/.test(name)
}
