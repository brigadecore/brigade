import { Job } from "./jobs"

/**
 * @deprecated Use Job.sequence or Job.parallel instead
 */
export class Group {

  /**
   * @deprecated Use Job.parallel followed by Job#run instead
   */
  public static async runAll(jobs: Job[]): Promise<void[]> {
    const group = new Group(jobs)
    return group.runAll()
  }

  /**
   * @deprecated Use Job.sequence followed by Job#run instead
   */
  public static async runEach(jobs: Job[]): Promise<void> {
    const group = new Group(jobs)
    return group.runEach()
  }

  private jobs: Job[]

  public constructor(jobs?: Job[]) {
    this.jobs = jobs || []
  }

  public add(...jobs: Job[]): void {
    for (const job of jobs) {
      this.jobs.push(job)
    }
  }

  public length(): number {
    return this.jobs.length
  }

  public async runEach(): Promise<void> {
    for (const job of this.jobs) {
      await job.run()
    }
  }

  public async runAll(): Promise<void[]> {
    const promises: Promise<void>[] = []
    for (const job of this.jobs) {
      promises.push(job.run())
    }
    return Promise.all(promises)
  }

}
