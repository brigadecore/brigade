import * as fs from "fs"
import * as moduleAlias from "module-alias"
import * as path from "path"
import * as requireFromString from "require-from-string"

import { Event } from "../../brigadier/src/events"
import { events, logger } from "./brigadier"

logger.info(`brigade-worker version: ${process.env.WORKER_VERSION}`)

const event: Event = require("/var/event/event.json") // eslint-disable-line @typescript-eslint/no-var-requires

logger.level = (event.worker.logLevel || "info").toLowerCase()

let script = ""
const scriptPath = path.join("/var/vcs", event.worker.configFilesDirectory, "brigade.js")
if (fs.existsSync(scriptPath)) {
  script = fs.readFileSync(scriptPath, "utf8")
} else {
  script = event.worker.defaultConfigFiles["brigade.js"]
}

if (!script) {
  logger.error("no brigade.js script found")
  process.exit(1)
}

// Install aliases for common ways of referring to Brigade/Brigadier.
moduleAlias.addAliases({
  "brigade": __dirname + "/brigadier",
  "brigadier": __dirname + "/brigadier",
  "@brigadecore/brigadier": __dirname + "/brigadier",
})

// Add the current module resolution paths to module-alias, so the
// node_modules that prestart.js adds to will be resolvable from the Brigade
// script and any local dependencies.
module.paths.forEach(moduleAlias.addPath)

moduleAlias()

requireFromString(script)

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

events.fire(event)
