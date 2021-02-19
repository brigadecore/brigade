import "mocha"
import { assert } from "chai"

import { Event } from "../src/events"
import { Job, Container, JobHost } from "../src/jobs"

import { MockJob } from "./mocks"

describe("jobs", () => {

  const event: Event = {
    id: "123456789",
    project: {
      id: "manhattan",
      secrets: {}
    },
    source: "foo",
    type: "bar",
    worker: {
      apiAddress: "",
      apiToken: "",
      configFilesDirectory: "",
      defaultConfigFiles: {}
    }
  }

  describe("Job", () => {
    describe("#constructor", () => {
      const job = Job.container("my-name", "debian:latest", event)
      it("initializes fields properly", () => {
        assert.equal(job.name, "my-name")
        assert.deepEqual(new Container("debian:latest"), job.primaryContainer)
        assert.deepEqual({}, job.sidecarContainers)
        assert.equal(1000 * 60 * 15, job.timeout)
        assert.deepEqual(new JobHost(), job.host)
      })
    })
  })

  describe("Container", () => {
    describe("#constructor", () => {
      const container = new Container("debian:latest")
      it("initializes fields properly", () => {
        assert.equal("debian:latest", container.image)
        assert.equal("IfNotPresent", container.imagePullPolicy)
        assert.deepEqual([], container.command)
        assert.deepEqual([], container.arguments)
        assert.deepEqual({}, container.environment)
        assert.isEmpty(container.workspaceMountPath)
        assert.isEmpty(container.sourceMountPath)
        assert.isFalse(container.privileged)
        assert.isFalse(container.useHostDockerSocket)
      })
    })
  })

  describe("JobHost", () => {
    describe("#constructor", () => {
      const jobHost = new JobHost()
      it("initializes fields properly", () => {
        assert.isUndefined(jobHost.os)
        assert.isDefined(jobHost.nodeSelector)
        assert.equal(0, Object.keys(jobHost.nodeSelector).length)
      })
    })
  })

  describe("SequentialJob", () => {
    it("runs sub-jobs in sequence", (done) => {
      const ledger: string[] = []
      const job0 = new MockJob("first", "debian:latest", event, () => {
        ledger.push("first")
      })
      const job1 = new MockJob("second", "debian:latest", event, () => {
        ledger.push("second")
      })
      const job2 = new MockJob("third", "debian:latest", event, () => {
        ledger.push("third")
      })
      const sequence = Job.sequence([job0, job1, job2])
      sequence.run().then(() => {
        assert.deepEqual(ledger, ["first", "second", "third"])
        done()
      })
    })
    it("stops processing on an error", (done) => {
      const ledger: string[] = []
      const job0 = new MockJob("first", "debian:latest", event, () => {
        ledger.push("first")
      })
      const job1 = new MockJob("second", "debian:latest", event, () => {
        ledger.push("second")
      })
      job1.fail = true
      const job2 = new MockJob("third", "debian:latest", event, () => {
        ledger.push("third")
      })
      Job.sequence([job0, job1, job2]).run().then(() => {
        done("expected error on job 1")
      }).catch(msg => {
        assert.equal(msg, "Failed")
        assert.equal(ledger.length, 2)  // job0 and job1 pushed, but then job1 failed
        done()
      })
    })
  })

  describe("ParallelJob", () => {
    it("runs jobs asynchronously", (done) => {
      const ledger: string[] = []
      const job0 = new MockJob("first", "debian:latest", event, () => {
        ledger.push("first")
      })
      job0.delay = 10
      const job1 = new MockJob("second", "debian:latest", event, () => {
        ledger.push("second")
      })
      job1.delay = 5
      const job2 = new MockJob("third", "debian:latest", event, () => {
        ledger.push("third")
      })
      job2.delay = 1
      Job.parallel([job0, job1, job2]).run().then(() => {
        // If these were executed concurrently, they should finish in reverse
        // order because of the specific delay values on each.
        assert.deepEqual(ledger, ["third", "second", "first"])
        done()
      })
    })
    it("stops processing on an error", (done) => {
      const job0 = new MockJob("first", "debian:latest", event, () => () => {
        // Do nothing
      })
      const job1 = new MockJob("second", "debian:latest", event, () => () => {
        // Do nothing
      })
      job1.fail = true
      const job2 = new MockJob("third", "debian:latest", event, () => () => {
        // Do nothing
      })
      Job.parallel([job0, job1, job2]).run().then(() => {
        done("expected error on job 2")
      }).catch(msg => {
        assert.equal(msg, "Failed")
        done()
      })
    })
  })

  describe("RetryableJob", () => {
    it("runs once if the job succeeds", (done) => {
      const ledger: number[] = []
      let index = 0
      const job = new MockJob("first", "debian:latest", event, () => {
        ledger.push(++index)
      })
      Job.retryable(job, 5).run().then(() => {
        assert.deepEqual(ledger, [1])
        done()
      })
    })
    it("retries on an error, until it succeeds", (done) => {
      const ledger: number[] = []
      let index = 0
      const job = new MockJob("first", "debian:latest", event, () => {
        ledger.push(++index)
        job.fail = index < 3
      })
      Job.retryable(job, 5).run().then(() => {
        assert.deepEqual(ledger, [1, 2, 3])
        done()
      })
    })
    it("stops retrying after maxAttempts", (done) => {
      const ledger: number[] = []
      let index = 0
      const job = new MockJob("first", "debian:latest", event, () => {
        ledger.push(++index)
      })
      job.fail = true
      Job.retryable(job, 5).run().then(() => {
        done("expected error")
      }).catch(msg => {
        assert.equal(msg, "Failed")
        assert.deepEqual(ledger, [1, 2, 3, 4, 5])
        done()
      })
    })
  })
})