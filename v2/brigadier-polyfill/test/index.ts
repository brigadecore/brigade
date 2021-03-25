import "mocha"
import { assert } from "chai"

import * as brigadier from "../src"

describe("brigadier", () => {
  it("has expected exports", () => {
    assert.property(brigadier, "Container")
    assert.property(brigadier, "ConcurrentGroup")
    // TODO: Figure out how to assert that an interface was exported.
    // assert.property(brigadier, "Event")
    assert.property(brigadier, "events")
    assert.property(brigadier, "Job")    
    assert.property(brigadier, "JobHost")
    assert.property(brigadier, "logger")
    assert.property(brigadier, "SerialGroup")
  })
})
