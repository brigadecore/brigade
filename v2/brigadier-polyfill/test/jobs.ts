import "mocha"
import { assert } from "chai"

import { Container, Event, Job, JobHost } from "../src"
import { ImagePullPolicy } from "../../brigadier/dist/jobs"

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
        assert.isDefined(job.logger)
      })
    })
  })

})
