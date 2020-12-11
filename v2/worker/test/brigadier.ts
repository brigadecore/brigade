import "mocha"
import { assert } from "chai"

import * as brigadier from "../src/brigadier"

describe("brigadier", () => {
  it("has expected exports", () => {
    assert.property(brigadier, "Container")
    assert.property(brigadier, "events")
    assert.property(brigadier, "Group")
    assert.property(brigadier, "Job")    
    assert.property(brigadier, "JobHost")
    assert.property(brigadier, "logger")
  })
})
