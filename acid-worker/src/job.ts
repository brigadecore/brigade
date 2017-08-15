/**
 * Package job provides support for jobs.
 *
 * A Job idescribes a particular unit of a build. A Job returns a Result.
 * A JobRunner is an implementation of the runtime logic for a Job.
 */

/** */

import {AcidEvent, Project} from "./events"

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
const acidImage: string = 'debian:jessie-slim'

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
  public image: string = acidImage

  /** Path to mount as the base path for executable code in the container.*/
  public mountPath: string = "/src"

  /** Set the max time to wait for this job to complete.*/
  public timeout: number = defaultTimeout

  /** Fetch the source repo. Default: true*/
  public useSource: boolean = true

  /** If true, the job will be run in privileged mode.
   * This is necessary for Docker builds.
   */
  public privileged: boolean = false

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
  constructor(name: string, image?: string, tasks?: string[]) {
    this.name = name
    this.image = image
    this.tasks = tasks || []
    this.env = {}
  }

  /** run executes the job and then */
  public abstract run(): Promise<Result>
}
