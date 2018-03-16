import "mocha";
import { assert } from "chai";
import * as mock from "./mock";

import * as events from "../src/events";

describe("events", function() {
  // Here, we just want to ensure that objects exported to brigadier are
  // available.
  it("has .BrigadeEvent", function() {
    assert.property(events, "BrigadeEvent");
  });
  it("has .Project", function() {
    assert.property(events, "Project");
  });
  it("has .EventRegistry", function() {
    assert.property(events, "EventRegistry");
  });
  describe("EventRegistry", function() {
    let er: events.EventRegistry;
    beforeEach(function() {
      er = new events.EventRegistry();
    });
    describe("#constructor", function() {
      it("registers 'ping' handler", function() {
        assert.isTrue(er.has("ping"));
      });
    });
    describe("#on", function() {
      it("registers an event handler", function() {
        er.on("my-event", (e: events.BrigadeEvent, p: events.Project) => {});
        assert.isTrue(er.has("my-event"));
      });
    });
    describe("#fire", function() {
      it("executes an event handler", function() {
        let fired = false;
        let ename = "my-event";
        let myEvent = mock.mockEvent();
        let myProj = mock.mockProject();
        myEvent.type = ename;
        er.on(ename, (e: events.BrigadeEvent, p: events.Project) => {
          fired = true;
        });
        er.fire(myEvent, myProj);
        assert.isTrue(fired);
      });
      context("when calling an event with no handler", function() {
        it("does not cause an error (does nothing)", function() {
          // We want this behavior because we don't want to force every brigade.js
          // to implement every possible event.
          let myEvent = mock.mockEvent();
          let myProj = mock.mockProject();
          myEvent.type = "no-such-event";
          er.fire(myEvent, myProj);
        });
      });
    });
  });
});
