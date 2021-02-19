import { Event } from "./events"

const defaultTimeout: number = 1000 * 60 * 15

/**
 * A Brigade job.
 * 
 * The factory methods on this type create jobs in various ways. Jobs do not
 * run immediately when created. You must call the Job#run method to start
 * the job running.
 */
export abstract class Job {
  /**
   * Runs the job.
   */
  public abstract run(): Promise<void>;

  /**
   * Specifies a job that runs in a container. By default, the container is simply
   * executed (via its default entry point). Set the ContinerJob#tasks property to run
   * specific commands in the container instead.
   * @param name The name of the job
   * @param image The container image to run in this job
   * @param event The event that triggered the job
   */
  public static container(name: string, image: string,  event: Event): ContainerJob {
    return new ContainerJob(name, image, event);
  }

  /**
   * Specifies a job that consists of sub-jobs running in parallel.
   * When run, all sub-jobs are started simultaneously (subject to scheduling
   * constraints). The parallel job completes when all sub-jobs have
   * completed.
   * @param jobs The jobs to be run in parallel
   */
  public static parallel(jobs: Job[]): ParallelJob {
    return new ParallelJob(jobs);
  }

  /**
   * Specifies a job that consists of sub-jobs running one after another.
   * A new sub-job is started only when the previous one completes.
   * The sequence completes when the last job has completed (or when any
   * job fails).
   * @param jobs The jobs to be run in sequence
   */
  public static sequence(jobs: Job[]): SequentialJob {
    return new SequentialJob(jobs);
  }

  /**
   * Specifies a job that can be retried. The job will be run and re-run
   * until either it succeeds or it has been tried `maxAttempts` times.
   * (If maxAttempts is zero or less the job will not be tried at all.)
   * @param job The job to be run and retried if necessary
   * @param maxAttempts The maximum number of times to attempt the job
   */
  public static retryable(job: Job, maxAttempts: number): RetryableJob {
    return new RetryableJob(job, maxAttempts);
  }
}

export class ContainerJob extends Job {
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
    super();
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

export class ParallelJob extends Job {
  constructor(private readonly jobs: Job[]) {
    super();
  }

  public async run(): Promise<void> {
    const promises = this.jobs.map((job) => job.run());
    await Promise.all(promises);
  }
}

export class SequentialJob extends Job {
  constructor(private readonly jobs: Job[]) {
    super();
  }

  public async run(): Promise<void> {
    for (const job of this.jobs) {
      await job.run();
    }
  }
}

export class RetryableJob extends Job {
  constructor(private readonly job: Job, private readonly maxAttempts: number) {
    super();
  }

  public async run(): Promise<void> {
    let attemptCount = 0;
    while (attemptCount < this.maxAttempts) {
      attemptCount++;
      try {
        await this.job.run();
        return;
      } catch (e) {
        if (attemptCount >= this.maxAttempts) {
          throw e;
        }
      }
    }
  }
}
