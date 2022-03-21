import "mocha"
import { assert } from "chai"

import { Event } from "../src/events"
import { SerialGroup, ConcurrentGroup } from "../src/groups"
import { MockJob } from "./mocks"

describe("groups", () => {
  describe("SerialGroup", () => {
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
    describe("#add", () => {
      it("adds a job", () => {
        const group = new SerialGroup()
        assert.equal(group.length(), 0)
        const job0 = new MockJob("hello", "debian:latest", event, () => {
          // Do nothing
        })
        const job1 = new MockJob("goodbye", "debian:latest", event, () => {
          // Do nothing
        })
        group.add(job0)
        group.add(job1)
        assert.equal(group.length(), 2)
      })
    })
    describe("#run", () => {
      it("runs each job in order", (done) => {
        const ledger: string[] = []
        const group = new SerialGroup()
        const job0 = new MockJob("first", "debian:latest", event, () => {
          ledger.push("first")
        })
        const job1 = new MockJob("second", "debian:latest", event, () => {
          ledger.push("second")
        })
        const job2 = new MockJob("third", "debian:latest", event, () => {
          ledger.push("third")
        })
        group.add(job0, job1, job2)
        group.run().then(() => {
          assert.deepEqual(ledger, ["first", "second", "third"])
          done()
        })
      })
      context("when job fails", () => {
        it("stops processing with an error", (done) => {
          const group = new SerialGroup()
          const job0 = new MockJob(
            "first",
            "debian:latest",
            event,
            () => () => {
              // Do nothing
            }
          )
          const job1 = new MockJob(
            "second",
            "debian:latest",
            event,
            () => () => {
              // Do nothing
            }
          )
          job1.fail = true
          const job2 = new MockJob("third", "debian:latest", event, () => {
            done("expected error on job 1")
          })
          group.add(job0, job1, job2)
          group.run().catch((msg) => {
            assert.equal(msg, "Failed")
            done()
          })
        })
      })
    })
  })

  describe("ConcurrentGroup", () => {
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
    describe("#add", () => {
      it("adds a job", () => {
        const group = new ConcurrentGroup()
        assert.equal(group.length(), 0)
        const job0 = new MockJob("hello", "debian:latest", event, () => {
          // Do nothing
        })
        const job1 = new MockJob("goodbye", "debian:latest", event, () => {
          // Do nothing
        })
        group.add(job0)
        group.add(job1)
        assert.equal(group.length(), 2)
      })
    })
    describe("#run", () => {
      it("runs jobs asynchronously", (done) => {
        const ledger: string[] = []
        const group = new ConcurrentGroup()
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
        group.add(job0, job1, job2)
        group.run().then(() => {
          // If these were executed concurrently, they should finish in reverse
          // order because of the specific delay values on each.
          assert.deepEqual(ledger, ["third", "second", "first"])
          done()
        })
      })
      context("when job fails", () => {
        const group = new ConcurrentGroup()
        it("stops processing with an error", (done) => {
          const job0 = new MockJob(
            "first",
            "debian:latest",
            event,
            () => () => {
              // Do nothing
            }
          )
          const job1 = new MockJob(
            "second",
            "debian:latest",
            event,
            () => () => {
              // Do nothing
            }
          )
          job1.fail = true
          const job2 = new MockJob(
            "third",
            "debian:latest",
            event,
            () => () => {
              // Do nothing
            }
          )
          group.add(job0, job1, job2)
          group
            .run()
            .then(() => {
              done("expected error on job 2")
            })
            .catch((msg) => {
              assert.equal(msg, "Failed")
              done()
            })
        })
      })
    })
  })
})
