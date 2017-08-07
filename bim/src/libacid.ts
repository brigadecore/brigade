import * as jobImpl from "./job"
import * as eventsImpl from "./events"
import {JobRunner} from "./k8s"

/*
 * This is the wrapper library for client-facing features of acid.
 *
 * The idea is that an acid.js file should be able to import libacid and have
 * all the common tools for working with Acid.
 */

// These are filled by the 'fire' event handler.
let currentEvent = null
let currentProject = null

export let events = new eventsImpl.EventRegistry()

export function fire(e: eventsImpl.AcidEvent, p: eventsImpl.Project) {
  currentEvent = e
  currentProject = p
  events.fire(e, p)
}

export class Job extends jobImpl.Job{
  run(): Promise<jobImpl.Result> {
    let jr = new JobRunner(this, currentEvent, currentProject)
    this._podName = jr.name
    return jr.run()
  }
}

export class Group {
  protected jobs: jobImpl.Job[] = []
  constructor(jobs?: jobImpl.Job[]) {
    this.jobs = jobs || []
  }
  add(...j: jobImpl.Job[]): void {
    for (let jj of j) {
      this.jobs.push(jj)
    }
  }
  length(): number {
    return this.jobs.length
  }
  // runEach runs each job in order and waits for every one to finish.
  runEach(): Promise<jobImpl.Result> {
    return this.jobs.reduce((p: Promise<jobImpl.Result>, j: jobImpl.Job) => {
      return p.then(() => j.run())
    }, Promise.resolve(new EmptyResult()))
  }
  // runAll runs all jobs in parallel and waits for them all to finish.
  runAll(): Promise<jobImpl.Result[]> {
    let plist: Promise<jobImpl.Result>[] = []
    for (let j of this.jobs) {
      plist.push(j.run())
    }
    return Promise.all(plist)
  }
}

class EmptyResult implements jobImpl.Result {
  toString() {return ""}
}
