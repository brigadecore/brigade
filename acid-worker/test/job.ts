import "mocha"
import {assert} from "chai"
import * as mock from "./mock"

import {Job, Result} from "../src/job"

describe("job", function() {
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
  })
})
