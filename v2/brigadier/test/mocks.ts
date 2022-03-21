import { setTimeout } from "timers"

import { Event } from "../src/events"
import { Job } from "../src/jobs"

// MockJob extends Job to make success or failure configurable. This allows us
// to force Job failures when a test case requires it.
export class MockJob extends Job {
  public fail = false
  public delay = 1 // Just enough to cause the event loop to sleep.
  public handler: () => void

  constructor(name: string, image: string, event: Event, handler: () => void) {
    super(name, image, event)
    this.handler = handler
  }

  public run(): Promise<void> {
    return new Promise((resolve, reject) => {
      setTimeout(() => {
        this.handler()
        if (this.fail) {
          reject("Failed")
        } else {
          resolve()
        }
      }, this.delay)
    })
  }
}
