import "mocha"
import {assert} from "chai"
import * as events from "../src/events"
import * as app from "../src/app"
import * as mock from "./mock"
import * as libacid from "../src/libacid"

app.setLoader((id: string, ns: string): Promise<events.Project> => {
  let proj =  mock.mockProject()
  proj.id = id
  proj.kubernetes.namespace = ns
  return Promise.resolve(proj)
})

describe("app", function() {
  describe("App", function() {
    let a: app.App
    let projectID: string = "app-test-id"
    let projectNS: string = "app-test-ns"
    beforeEach(function() {
      a = new app.App(projectID, projectNS)
      // Disable this so we can run tests without panics.
      a.exitOnError = false
    })
    describe("App.generateBuildID", function() {
      // acid-worker-01BR5VRVP06Q0BASBVB1WYK7X1-01234567
      let commit = "01234567890"
      assert.match(app.App.generateBuildID(commit), /acid-worker-[A-Z0-9]{26}-01234567/)
    })
    describe("#run", function() {
      it("runs an event handler to completion", function(done) {
        let e = mock.mockEvent()
        e.type = "ping"
        a.run(e)
        done()
      })
      context("when no event handler is registered", function() {
        it("silently completes", function(done) {
          let e = mock.mockEvent()
          e.type = "no such event"
          a.run(e)
          done()
        })
      })
      context("when 'after' event handler is registered", function() {
        it("calls after handler", function(done) {
          let after = 1
          libacid.events.on("test-after", () => {
            after++
          })
          libacid.events.on("after", () => {
            after++
          })
          let e = mock.mockEvent()
          e.type = "test-after"
          a.run(e)
          setTimeout(() => {
            assert.equal(3, after)
            done()
          }, 10)
        })
      })
      context("when an event handler emits an uncaught rejection", function() {
        it("calls error event", function(done) {
          libacid.events.on("test-fail", () => {
            Promise.reject("intentional error")
          })
          let caught = false
          libacid.events.on("error", () => {
            caught = true
          })
          let e = mock.mockEvent()
          e.type = "test-fail"
          a.run(e)
          setTimeout(() => {
            assert.isTrue(caught)
            done()
          }, 10)
        })
      })
      context("when a job throws an exception", function() {
        it("calls error event", function(done) {
          libacid.events.on("test-fail", () => {
            throw "can't touch this"
          })
          let caught = false
          libacid.events.on("error", () => {
            caught = true
          })
          let e = mock.mockEvent()
          e.type = "test-fail"
          a.run(e)
          setTimeout(() => {
            assert.isTrue(caught)
            done()
          }, 10)
        }) // turtles
      }) // all
    }) // the
  }) // way
}) // down
