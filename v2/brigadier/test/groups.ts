import "mocha"
import { assert } from "chai"

import { Event } from "../src/events"
import { Group } from "../src/groups"
import { MockJob } from "./mocks"

describe("groups", () => {

  describe("Group", () => {
    const event: Event = {
      id: "123456789",
      project: {
        id: "manhattan",
        secrets: new Map<string, string>()
      },
      source: "foo",
      type: "bar",
      worker: {
        apiAddress: "",
        apiToken: "",
        configFilesDirectory: "",
        defaultConfigFiles: new Map<string, string>()
      }
    }
    describe("#add", () => {
      it("adds a job", () => {
        const group = new Group()
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
    describe("#runEach", () => {
      it("runs each job in order", (done) => {
        const ledger: string[] = []
        const group = new Group()
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
        group.runEach().then(() => {
          assert.deepEqual(ledger, ["first", "second", "third"])
          done()
        })
      })
      context("when job fails", () => {
        it("stops processing with an error", (done) => {
          const group = new Group()
          const job0 = new MockJob("first", "debian:latest", event, () => () => {
            // Do nothing
          })
          const job1 = new MockJob("second", "debian:latest", event, () => () => {
            // Do nothing
          })
          job1.fail = true
          const job2 = new MockJob("third", "debian:latest", event, () => {
            done("expected error on job 1")
          })
          group.add(job0, job1, job2)
          group.runEach().catch(msg => {
            assert.equal(msg, "Failed")
            done()
          })
        })
      })
    })
    describe("#runAll", () => {
      it("runs jobs asynchronously", (done) => {
        const ledger: string[] = []
        const group = new Group()
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
        group.runAll().then(() => {
          // If these were executed concurrently, they should finish in reverse
          // order because of the specific delay values on each.
          assert.deepEqual(ledger, ["third", "second", "first"])
          done()
        })
      })
      context("when job fails", () => {
        const group = new Group()
        it("stops processing with an error", (done) => {
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
          group.add(job0, job1, job2)
          group.runAll().then(() => {
            done("expected error on job 2")
          }).catch(msg => {
            assert.equal(msg, "Failed")
            done()
          })
        })
      })
    })
    describe("static #runEach", () => {
      it("runs each job in order", (done) => {
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
        Group.runEach([job0, job1, job2]).then(() => {
          assert.deepEqual(ledger, ["first", "second", "third"])
          done()
        })
      })
      context("when job fails", () => {
        it("stops processing with an error", (done) => {
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
          Group.runEach([job0, job1, job2]).then(() => {
            done("expected error on job 2")
          }).catch(msg => {
            assert.equal(msg, "Failed")
            done()
          })
        })
      })
    })
    describe("static #runAll", () => {
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
        Group.runAll([job0, job1, job2]).then(() => {
          // If these were executed concurrently, they should finish in reverse
          // order because of the specific delay values on each.
          assert.deepEqual(ledger, ["third", "second", "first"])
          done()
        })
      })
      context("when job fails", () => {
        it("stops processing with an error", (done) => {
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
          Group.runAll([job0, job1, job2]).then(() => {
            done("expected error on job 1")
          }).catch(msg => {
            assert.equal(msg, "Failed")
            done()
          })
        })
      })
    })
  })

})