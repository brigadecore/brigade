import { Runnable } from "./runnables"

/**
 * The base type for Runnables composed of other Runnables.
 * Do not construct the base Group type; use Job.sequential (or SerialGroup)
 * or Job.concurrent (or ConcurrentGroup) instead.
 * 
 * @abstract
 */
class Group {
  protected runnables: Runnable[]

  protected constructor(...runnables: Runnable[]) {
    this.runnables = runnables || []
  }

  public add(...runnables: Runnable[]): void {
    for (const runnable of runnables) {
      this.runnables.push(runnable)
    }
  }

  public length(): number {
    return this.runnables.length
  }
}

/**
 * A Runnable composed of sub-Runnables (such as jobs
 * or concurrent groups) running one after another.
 * A new Runnable is started only when the previous one completes.
 * The sequence completes when the last Runnable has completed (or when any
 * Runnable fails).
 * 
 * @param runnables The work items to be run in sequence
 */
export class SerialGroup extends Group implements Runnable {
  public constructor(...runnables: Runnable[]) {
    super(...runnables)
  }

  /**
   * Runs the serial group.
   * 
   * @returns A Promise which completes when the last item in the group completes (or
   * any item fails). If all items ran successfully, the Promise resolves; if any
   * item failed (that is, its Runnable#run Promise rejected), the Promise
   * rejects.
   */
  public async run(): Promise<void> {
    for (const runnable of this.runnables) {
      await runnable.run()
    }
  }

}

/**
 * A Runnable composed of sub-Runnables (such as jobs
 * or sequential groups) running concurrently.
 * When run, all Runnables are started simultaneously (subject to
 * scheduling constraints).
 * The concurrent group completes when all Runnables have completed.
 * 
 * @param runnables The work items to be run in parallel
 */
export class ConcurrentGroup extends Group implements Runnable {
  public constructor(...runnables: Runnable[]) {
    super(...runnables)
  }

  /**
   * Runs the concurrent group.
   * 
   * @returns A Promise which completes when all items in the group complete.
   * If all items ran successfully, the Promise resolves; if any
   * item failed (that is, its Runnable#run Promise rejected), the Promise
   * rejects.
   */
  public async run(): Promise<void> {
    const promises: Promise<void>[] = []
    for (const runnable of this.runnables) {
      promises.push(runnable.run())
    }
    try {
      await Promise.all(promises)
      return Promise.resolve()
    } catch(e) {
      return Promise.reject(e)
    }
  }

}
