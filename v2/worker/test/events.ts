import "mocha"
import { assert } from "chai"

import { events } from "../src/events"

describe("events", () => {

  describe("events", () => {
    let handlerCalled: boolean
    events.on("foo", "bar", () => {
      handlerCalled = true
    })
    describe("#fire", () => {
      beforeEach(() => {
        handlerCalled = false
      })
      context("when the handler is found in the map", () => {  
        it("invokes the handler", () => {
          events.fire({
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
          })
          assert.isTrue(handlerCalled)
        })
      })
      context("when the handler is not found in the map", () => {
        it("does not invoke any handler", () => {
          events.fire({
            id: "123456789",
            project: {
              id: "manhattan",
              secrets: {}
            },
            source: "bat",
            type: "baz",
            worker: {
              apiAddress: "",
              apiToken: "",
              configFilesDirectory: "",
              defaultConfigFiles: {}
            }
          })
          assert.isFalse(handlerCalled)
        })
      })
    })
  })

})
