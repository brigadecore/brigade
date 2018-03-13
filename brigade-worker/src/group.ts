/**
 * group provides features for grouping jobs and managing them collectively.
 */

/** */

import * as jobImpl from "./job";
import {ContextLogger} from './logger';

const logger = new ContextLogger("group");

/**
 * Group describes a collection of associated jobs.
 */
export class Group {
  /**
   * runAll is a convenience for running jobs in parallel.
   *
   * This runs a series of jobs in parallel. It is equivalent to
   * `(new Group(jobs)).runAll()`
   */
  public static runAll(jobs: jobImpl.Job[]): Promise<jobImpl.Result[]> {
    let g = new Group(jobs);
    return g.runAll();
  }
  /**
   * runEach is a convenience of running jobs in a sequence.
   *
   * This runs a series of jobs in order, blocking on each until it completes.
   * It is equivalent to `(new Group(jobs)).runEach()`
   */
  public static runEach(jobs: jobImpl.Job[]): Promise<jobImpl.Result[]> {
    let g = new Group(jobs);
    return g.runEach();
  }

  protected jobs: jobImpl.Job[] = [];
  public constructor(jobs?: jobImpl.Job[]) {
    this.jobs = jobs || [];
  }
  /**
   * add adds one or more jobs to the group.
   */
  public add(...j: jobImpl.Job[]): void {
    for (let jj of j) {
      this.jobs.push(jj);
    }
  }
  /**
   * length returns the number of items in the group
   */
  public length(): number {
    return this.jobs.length;
  }
  /**
   * runEach runs each job in order and waits for every one to finish.
   */
  public runEach(): Promise<jobImpl.Result[]> {
    return this.jobs.reduce(
      (promise: Promise<jobImpl.Result[]>, job: jobImpl.Job) => {
        return promise.then((results: jobImpl.Result[]) => {
          return job.run().then(jobResult => {
            results.push(jobResult);
            return results;
          });
        });
      },
      Promise.resolve([])
    );
  }
  /**
   * runAll runs all jobs in parallel and waits for them all to finish.
   */
  public runAll(): Promise<jobImpl.Result[]> {
    let plist: Promise<jobImpl.Result>[] = [];
    for (let j of this.jobs) {
      plist.push(j.run());
    }
    return Promise.all(plist);
  }
  /**
   * batchRunAll runs all jobs in parallel in batches of up to <maxConcurrent> and waits for them all to finish.
   */
  public batchRunAll(maxConcurrent: number): Promise<jobImpl.Result[]> {
    const jobs = this.jobs;
    return new Promise(function(resolve, reject) {
      const results = [];
      let i = 0;

      function next() {
        if (i < jobs.length) {
          let plist: Promise<jobImpl.Result>[] = [];
          logger.log("Running jobs from", i, "to", i + maxConcurrent);
          for (let j of jobs.slice(i, i + maxConcurrent)) {
            plist.push(j.run());
          }
          Promise.all(plist).then(function(data) {
            results.push(data);
            next();
          }, reject);
          i += maxConcurrent;
        } else {
          resolve(results);
        }
      }
      next();
    });
  }
}

/**
 * EmptyResults is an empty Result object.
 */
class EmptyResult implements jobImpl.Result {
  toString() {
    return "";
  }
}
