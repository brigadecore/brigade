import "mocha"
import {assert} from "chai"
import * as mock from "./mock"

import {Job, Result, JobCache, JobStorage, acidCachePath, acidStoragePath} from "../src/job"

describe("job", function() {
  describe("JobCache", function() {
    describe("#constructor", function() {
      it("correctly sets default values", function(){
        let c = new JobCache()
        assert.equal(c.path, acidCachePath, "Dir is /acid/cache")
        assert.isFalse(c.enabled, "disabled by default")
        assert.equal(c.size, "5Mi", "size is 5mi")
      })
    })
  })
  describe("JobStorage", function() {
    describe("#constructor", function() {
      it("correctly sets default values", function(){
        let c = new JobStorage()
        assert.equal(c.path, acidStoragePath, "Dir is " + acidStoragePath)
        assert.isTrue(c.enabled, "enabled by default")
      })
    })
  })
  describe("Job", function() {
    let j: mock.MockJob
    describe("#constructor", function() {
      it("creates a named job", function() {
        j = new mock.MockJob("myName")
        assert.equal(j.name, "myName")
      })
      context("when image is supplied", function() {
        it("sets image property", function() {
          j = new mock.MockJob("myName", "alpine:3.4")
          assert.equal(j.image, "alpine:3.4")
        })
      })
      context("when tasks are supplied", function() {
        it("sets task list", function() {
          j = new mock.MockJob("my", "img", ["a", "b", "c"])
          assert.deepEqual(j.tasks, ["a", "b", "c"])
        })
      })

    })
    describe("#podName", function() {
      beforeEach(function(){
        j = new mock.MockJob("my-job")
      })
      context("before run", function() {
        it("is empty", function() {
          assert.isUndefined(j.podName)
        })
      })
      context("after run", function() {
        it("is accessible", function(done) {
          j.run().then((rez) => {
            assert.equal(j.podName, "generated-fake-job-name")
            done()
          })
        })
      })
    })
    describe("#cache", function() {
      it("is disabled by default", function() {
        assert.isFalse(j.cache.enabled)
      })
    })
    describe("#storage", function() {
      it("is enabled by default", function() {
        assert.isTrue(j.storage.enabled)
      })
    })
  })
})
