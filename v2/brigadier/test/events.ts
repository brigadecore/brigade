import "mocha"
import { assert } from "chai"

import { EventRegistry, EventHandler } from "../src/events"

describe("events", () => {

  describe("EventRegistry", () => {
    describe("#on", () => {
      // We cannot see directly into EventRegistry's protected internal map of
      // handlers to assert it is managed correctly, but we CAN extend
      // EventRegistry and add an accessor so that we can get at handlers.
      class ER extends EventRegistry {
        public getHandler(source: string, type: string): EventHandler | undefined {
          return this.handlers.get(`${source}:${type}`)
        }
      }
      it("adds the handler to the map", () => {
        const e = new ER()
        const handler = () => {
          // Do nothing
        }
        e.on("foo", "bar", handler)
        assert.equal(handler, e.getHandler("foo", "bar"))
      })
    })
  })

})
