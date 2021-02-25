import { Event } from "./events"
import { ConcurrentGroup, SerialGroup } from "./groups"
import { Runnable } from "./runnables"

const defaultTimeout: number = 1000 * 60 * 15

/**
 * A Brigade job.
 * 
 * Instances of Job represent containers that your Brigade script can
 * run using the Job#run method. By default, the container is simply
 * executed (via its default entry point). Set the Job#primaryContainer#command
 * property to run specific commands in the container instead.
 * 
 * Job also provides static methods for building up runnable
 * elements out of more basic ones. For example, to compose
 * a set of workloads into a sequence, you can use the
 * Job.sequence method.
 */
export class Job implements Runnable {
  public name: string
  public primaryContainer: Container
  public sidecarContainers: { [key: string]: Container } = {}
  public timeout: number = defaultTimeout
  public host: JobHost = new JobHost()
  protected event: Event

  constructor(
    name: string,
    image: string,
    event: Event
  ) {
    this.name = name
    this.primaryContainer = new Container(image)
    this.event = event
  }

  public run(): Promise<void> {
    return Promise.resolve()
  }

  public logs(): Promise<string> {
    return Promise.resolve("skipped logs")
  }

  /**
   * Creates a Job.
   * 
   * (Note: This is equivalent to `new Job(...)`. It is provided so that
   * script authors have the option of a consistent style for creating
   * and composing jobs and groups.)
   * 
   * @param name The name of the job
   * @param image The OCI image reference for the primary container
   * @param event The event that triggered the job
   */
  public static container(
    name: string,
    image: string,
    event: Event
  ): Job {
    return new Job(name, image, event)
  }

  /**
   * Specifies a Runnable that consists of sub-Runnables (such as jobs
   * or concurrent groups) running one after another.
   * A new Runnable is started only when the previous one completes.
   * The sequence completes when the last Runnable has completed (or when any
   * Runnable fails).
   * 
   * @param runnables The work items to be run in sequence
   */
  public static sequence(...runnables: Runnable[]): SerialGroup {
    return new SerialGroup(...runnables)
  }

  /**
   * Specifies a Runnable that consists of sub-Runnables (such as jobs
   * or sequential groups) running concurrently.
   * When run, all Runnables are started simultaneously (subject to
   * scheduling constraints).
   * The concurrent group completes when all Runnables have completed.
   * 
   * @param runnables The work items to be run in parallel
   */
  public static concurrent(...runnables: Runnable[]): ConcurrentGroup {
    return new ConcurrentGroup(...runnables)
  }
}

export class Container {
  public image: string
  public imagePullPolicy = "IfNotPresent"
  public workingDirectory = ""
  public command: string[] = []
  public arguments: string[] = []
  public environment: { [key: string]: string } = {}
  public workspaceMountPath = ""
  public sourceMountPath = ""
  public privileged = false
  public useHostDockerSocket = false

  constructor(image: string) {
    this.image = image
  }
}

export class JobHost {
  public os?: string
  public nodeSelector: { [key: string]: string } = {}
}
