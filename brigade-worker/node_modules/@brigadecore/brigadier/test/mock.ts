import { Project, BrigadeEvent } from "../src/events";
import { Result, Job } from "../src/job";

// This package contains mocks of objects found elsewhere in Brigade.

export function mockProject(): Project {
  return {
    id: "brigade-4897c99315be5d2a2403ea33bdcb24f8116dc69613d5917d879d5f",
    name: "brigadecore/empty-testbed",
    repo: {
      name: "brigadecore/empty-testbed",
      cloneURL: "https://github.com/brigadecore/empty-testbed.git",
      token: "supersecret",
      initGitSubmodules: false
    },
    kubernetes: {
      namespace: "default",
      vcsSidecar: "brigadecore/git-sidecar:latest",
      buildStorageSize: "50Mi"
    },
    allowPrivilegedJobs: true,
    allowHostMounts: false
  } as Project;
}

export function mockEvent() {
  return {
    buildID: "1234567890abcdef",
    workerID: "test-1234567890abcdef-12345678",
    type: "push",
    provider: "github",
    revision: {
      commit: "c0ffee"
    },
    payload: "{}"
  } as BrigadeEvent;
}

export class MockResult implements Result {
  protected msg: string = "uninitialized";
  constructor(msg: string) {
    this.msg = msg;
  }
  public toString(): string {
    return this.msg;
  }
}

// MockJob implements the run() method on Job with a resolved Promise<MockResult>.
//
// If 'MockJob.fail = true', the job will return a failure instead of a success.
//
// The MockJob.run method will sleep for one nanosecond (that is, give up at least
// one scheduler run). To set a longer delay, set MockJob.delay.
export class MockJob extends Job {
  public fail: boolean = false;
  public delay: number = 1; // Just enough to cause the event loop to sleep it.
  public run(): Promise<Result> {
    let fail = this.fail;
    let delay = this.delay;
    this._podName = "generated-fake-job-name";
    return new Promise((resolve, reject) => {
      if (fail) {
        setTimeout(() => {
          reject("Failed");
        }, delay);
        return;
      }
      setTimeout(() => {resolve(new MockResult(this.name))}, delay);
    });
  }
  public logs(): Promise<string> {
    let fail = this.fail;
    let delay = this.delay;
    this._podName = "generated-fake-job-name-2";
    return new Promise((resolve, reject) => {
      if (fail) {
        setTimeout(() => {
          resolve(`These are the logs showing failure.`);
        }, delay);
        return;
      }
      setTimeout(() => {
        resolve(`These are the logs showing successful completion.`);
      }, delay);
    });
  }
}

export class MockBuildStorage {
  public create(
    e: BrigadeEvent,
    project: Project,
    size?: string
  ): Promise<string> {
    return Promise.resolve(e.workerID || "unknown");
  }
  public destroy(): Promise<boolean> {
    return Promise.resolve(true);
  }
}
