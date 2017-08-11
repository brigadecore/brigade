import "mocha"
import {assert} from "chai"

import * as acid from "../src/libacid"
import * as jobImpl from "../src/job"

import * as mock from "./mock"

// These tests are largely designed to ensure that the objects a script is likely
// to use are indeed exposed. Tests for the actual functionality of each are found
// in their respective libraries.
describe("libacid", function() {
  it("has #fire", function() {
    assert.property(acid, "fire")
  })
  it("has .Job", function() {
    assert.property(acid, "Job")
  })
  it("has .Group", function() {
    assert.property(acid, "Group")
  })
  it("has .events", function() {
    assert.property(acid, "events")
  })

  // Events tests
  describe("events", function() {
    it("has #on", function() {
      assert.property(acid.events, "on")
    })
  })

  // Group tests
  describe("Group", function() {
    let g: acid.Group
    beforeEach(function() {
      g = new acid.Group()
    })
    describe("#add", function() {
      it("adds a job", function() {
        assert.equal(g.length(), 0)
        let j = new mock.MockJob("hello")
        let j2 = new mock.MockJob("goodbye")
        g.add(j)
        g.add(j2)
        assert.equal(g.length(), 2)
      })
    })
    describe("#runEach", function() {
      it("runs each job in order", function(done) {
        let j1 = new mock.MockJob("first")
        let j2 = new mock.MockJob("second")
        let j3 = new mock.MockJob("third")
        // This ensures that if the jobs were not executed in sequence,
        // 1 and 2 would finish before 3.
        j3.delay = 50
        g.add(j1, j2, j3)
        g.runEach().then((rez: jobImpl.Result) => {
          assert.equal(rez.toString(), j3.name)
          done()
        })
      })
      context("when job fails", function() {
        it("stops processing with an error", function(done) {
          let j1 = new mock.MockJob("first")
          let j2 = new mock.MockJob("second")
          j2.fail = true
          let j3 = new mock.MockJob("third")
          g.add(j1, j2, j3)
          g.runEach().then((rez: jobImpl.Result) => {
            done("expected error on job 2")
          }).catch((msg) => {
            assert.equal(msg, "Failed")
            done()
          })
        })
      })
    })
    describe("#runAll", function() {
      it("runs jobs asynchronously", function(done) {
        let j1 = new mock.MockJob("first")
        let j2 = new mock.MockJob("second")
        let j3 = new mock.MockJob("third")
        g.add(j1, j2, j3)
        g.runAll().then((rez: jobImpl.Result[]) => {
          assert.equal(rez.length, 3)
          done()
        })
      })
      context("when job fails", function() {
        it("stops processing with an error", function(done) {
          let j1 = new mock.MockJob("first")
          let j2 = new mock.MockJob("second")
          j2.fail = true
          let j3 = new mock.MockJob("third")
          g.add(j1, j2, j3)
          g.runAll().then((rez: jobImpl.Result) => {
            done("expected error on job 2")
          }).catch((msg) => {
            assert.equal(msg, "Failed")
            done()
          })
        })
      })
    })
  })
})
