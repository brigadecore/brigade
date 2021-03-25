import { Event, EventRegistry as BrigadierEventRegistry } from "@brigadecore/brigadier"

import { logger } from "./logger"

class EventRegistry extends BrigadierEventRegistry {

  public process(): void {
    const event: Event = require("/var/event/event.json") // eslint-disable-line @typescript-eslint/no-var-requires

    logger.level = (event.worker.logLevel || "info").toLowerCase()

    let exitCode = 0

    process.on("unhandledRejection", (reason) => {
      logger.error(reason)
      exitCode = 1
    })

    process.on("exit", code => {
      if (code != 0) {
        process.exit(code)
      }
      if (exitCode != 0) {
        process.exit(exitCode)
      }
    })

    this.fire(event)
  }

}

export const events = new EventRegistry()
