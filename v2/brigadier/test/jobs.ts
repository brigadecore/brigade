import "mocha"
import { assert } from "chai"

import { Event } from "../src/events"
import { Job, Container, JobHost } from "../src/jobs"

describe("jobs", () => {

  describe("Job", () => {
    describe("#constructor", () => {
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
      const job = new Job("my-name", "debian:latest", event)
      it("initializes fields properly", () => {
        assert.equal(job.name, "my-name")
        assert.deepEqual(new Container("debian:latest"), job.primaryContainer)
        assert.deepEqual(new Map<string, Container>(), job.sidecarContainers)
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
        assert.deepEqual(new Map<string, string>(), container.environment)
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
        assert.equal(0, jobHost.nodeSelector.size)
      })
    })
  })

})