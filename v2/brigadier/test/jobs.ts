import "mocha"
import { assert } from "chai"

import { Event } from "../src/events"
import { Job, Container, JobHost, ImagePullPolicy } from "../src/jobs"

describe("jobs", () => {

  describe("Job", () => {
    describe("#constructor", () => {
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
      const job = new Job("my-name", "debian:latest", event)
      it("initializes fields properly", () => {
        assert.equal(job.name, "my-name")
        assert.deepEqual(job.primaryContainer, new Container("debian:latest"))
        assert.deepEqual(job.primaryContainer.imagePullPolicy, ImagePullPolicy.IfNotPresent)
        assert.deepEqual(job.sidecarContainers, {})
        assert.equal(job.timeoutSeconds, 60 * 15)
        assert.deepEqual(job.host, new JobHost())
      })
    })
  })

  describe("Container", () => {
    describe("#constructor", () => {
      const container = new Container("debian:latest")
      it("initializes fields properly", () => {
        assert.equal(container.image, "debian:latest")
        assert.equal(container.imagePullPolicy, ImagePullPolicy.IfNotPresent)
        assert.deepEqual(container.command, [])
        assert.deepEqual(container.arguments, [])
        assert.deepEqual(container.environment, {})
        assert.isEmpty(container.workspaceMountPath)
        assert.isEmpty(container.sourceMountPath)
        assert.isFalse(container.privileged)
        // assert.isFalse(container.useHostDockerSocket)
      })
    })
  })

  describe("JobHost", () => {
    describe("#constructor", () => {
      const jobHost = new JobHost()
      it("initializes fields properly", () => {
        assert.isUndefined(jobHost.os)
        assert.isDefined(jobHost.nodeSelector)
        assert.equal(Object.keys(jobHost.nodeSelector).length, 0)
      })
    })
  })

})