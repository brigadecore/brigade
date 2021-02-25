/**
 * Any piece of work that can be run from a Brigade script.
 * The most fundamental Runnable is the Job class; other Runnables
 * are typically built up from Job instances.
 */
export interface Runnable {
  /**
   * Runs the workload specified by the Runnable instance.
   */
  run(): Promise<void>
}
