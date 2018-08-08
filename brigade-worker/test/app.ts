import "mocha";
import { assert } from "chai";
import * as events from "@azure/brigadier/out/events";
import * as app from "../src/app";
import * as mock from "./mock";
import * as brigadier from "../src/brigadier";

let loader = (id: string, ns: string): Promise<events.Project> => {
  let proj = mock.mockProject();
  proj.id = id;
  proj.kubernetes.namespace = ns;
  return Promise.resolve(proj);
};

describe("app", function() {
  describe("App", function() {
    let a: app.App;
    let projectID: string = "app-test-id";
    let projectNS: string = "app-test-ns";
    beforeEach(function() {
      a = new app.App(projectID, projectNS);
      a.loadProject = loader;
      a.buildStorage = new mock.MockBuildStorage();
      // Disable this so we can run tests without panics.
      a.exitOnError = false;
    });
    describe("#run", function() {
      it("runs an event handler to completion", function(done) {
        let e = mock.mockEvent();
        e.type = "ping";
        a.run(e);
        done();
      });
      context("when no event handler is registered", function() {
        it("silently completes", function(done) {
          let e = mock.mockEvent();
          e.type = "no such event";
          a.run(e);
          done();
        });
      });
      context("when an event handler emits an uncaught rejection", function() {
        it("calls error event", function(done) {
          brigadier.events.on("test-fail", () => {
            Promise.reject("intentional error");
          });
          let caught = false;
          brigadier.events.on("error", () => {
            caught = true;
          });
          let e = mock.mockEvent();
          e.type = "test-fail";
          a.run(e);
          setTimeout(() => {
            assert.isTrue(caught);
            done();
          }, 10);
        });
      });
      context("when a job throws an exception", function() {
        it("calls error event", function(done) {
          brigadier.events.on("test-fail", () => {
            throw "can't touch this";
          });
          let caught = false;
          brigadier.events.on("error", () => {
            caught = true;
          });
          let e = mock.mockEvent();
          e.type = "test-fail";
          a.run(e);
          setTimeout(() => {
            assert.isTrue(caught);
            done();
          }, 10);
        }); // turtles
      }); // all
    }); // the
  }); // way
}); // down
