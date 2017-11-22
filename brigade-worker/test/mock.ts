import {Project, BrigadeEvent} from "../src/events"
import {Result, Job} from "../src/job"


// This package contains mocks of objects found elsewhere in Brigade.


export function mockProject(): Project {
  return {
    id: "brigade-c0ff33544b459e6ac0ffee",
    name: "deis/empty-testbed",
    repo: {
      name: "deis/empty-testbed",
      cloneURL: "https://github.com/deis/empty-testbed.git",
      token: "supersecret"
    },
    kubernetes: {
      namespace: "default",
      vcsSidecar: "deis/git-sidecar:latest"
    },
    allowPrivilegedJobs: true,
    allowHostMounts: false,
  } as Project
}

export function mockEvent() {
  return {
    buildID: "1234567890abcdef",
    workerID: "test-1234567890abcdef-12345678",
    type: "push",
    provider: "github",
    commit: "c0ffee",
    payload: "{}"
  } as BrigadeEvent
}

export class MockResult implements Result {
  protected msg: string = "uninitialized"
  constructor(msg: string) {
    this.msg = msg
  }
  public toString(): string {
    return this.msg
  }
}

// MockJob implements the run() method on Job with a resolved Promise<MockResult>.
//
// If 'MockJob.fail = true', the job will return a failure instead of a success.
//
// The MockJob.run method will sleep for one nanosecond (that is, give up at least
// one scheduler run). To set a longer delay, set MockJob.delay.
export class MockJob extends Job {
  public fail: boolean = false
  public delay: number = 1 // Just enough to cause the event loop to sleep it.
  public run(): Promise<Result> {
    let fail = this.fail
    let delay = this.delay
    this._podName = "generated-fake-job-name"
    return new Promise((resolve, reject) => {
      if (fail) {
        setTimeout(() => {reject("Failed")}, delay)
        return
      }
      setTimeout(resolve(new MockResult(this.name)), delay)
    })
  }
}

export class MockBuildStorage {
  public create(e: BrigadeEvent, project: Project, size?: string): Promise<string> {
    return Promise.resolve(e.workerID)
  }
  public destroy(): Promise<boolean> {
    return Promise.resolve(true)
  }
}
