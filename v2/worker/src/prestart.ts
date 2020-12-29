import * as fs from "fs"
import { execFileSync } from "child_process"
import * as path from "path"

import { logger } from "../../brigadier/src"

const event = require("/var/event/event.json") // eslint-disable-line @typescript-eslint/no-var-requires

logger.level = (event.worker.logLevel || "info").toLowerCase()

const configFilePath = path.join("/var/vcs", event.worker.configFilesDirectory, "brigade.json")

let dependencies
if (fs.existsSync(configFilePath)) {
  dependencies = require(configFilePath).dependencies // eslint-disable-line @typescript-eslint/no-var-requires
} else if (event.worker.defaultConfigFiles) {
  const configFileContents = event.worker.defaultConfigFiles["brigade.json"]
  if (configFileContents) {
    dependencies = JSON.parse(configFileContents).dependencies
  } else {
    logger.debug("prestart: no dependencies file found")
    process.exit(0)
  }
}

if (!dependencies || Object.keys(dependencies).length == 0) {
  console.debug("prestart: no dependencies to install")
  process.exit(0)
}

if (require.main === module) { // This helps us NOT actually install packages while testing
  const packages = Object.entries(dependencies).map(([dep, version]) => dep + "@" + version)
  console.info(`installing ${packages.join(", ")}`)
  execFileSync("yarn", ["add", ...packages])
}
