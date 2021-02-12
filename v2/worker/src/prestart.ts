import * as fs from "fs"
import { execFileSync } from "child_process"
import * as path from "path"

import { logger } from "./logger"

const event = require("/var/event/event.json") // eslint-disable-line @typescript-eslint/no-var-requires

logger.level = (event.worker.logLevel || "info").toLowerCase()

const configFilesPath = path.join("/var/vcs", event.worker.configFilesDirectory)

// Create the configFilesPath if it doesn't already exist
if (!fs.existsSync(configFilesPath)) {
  fs.mkdirSync(configFilesPath, { recursive: true })
}

// If applicable, write defaultConfigFiles into configFilesPath
if (event.worker.defaultConfigFiles) {
  for (const filename in event.worker.defaultConfigFiles) {
    const fullFilePath = path.join(configFilesPath, filename)
    if (fs.existsSync(fullFilePath)) {
      logger.warn(`${fullFilePath} already exists; refusing to overwrite it with default ${filename}`)
    } else {
      logger.debug(`writing default ${filename} to ${fullFilePath}`)
      fs.writeFileSync(fullFilePath, event.worker.defaultConfigFiles[filename])
    }
  }
}

const brigadierPackageName = "@brigadecore/brigadier"
const brigadierSrcPath = "/var/brigade-worker/src/brigadier"
const packageJSONPath = path.join(configFilesPath, "package.json")
let pkg
if (fs.existsSync(packageJSONPath)) {
  logger.debug(`found an existing package.json at ${packageJSONPath}`)
  pkg = require(packageJSONPath)
  logger.debug(`patching package.json to use ${brigadierPackageName} included in worker image`)
  if (pkg.devDependencies) {
    delete pkg.devDependencies[brigadierPackageName]
  }
  if (pkg.peerDependencies) {
    delete pkg.peerDependencies[brigadierPackageName]
  }
  if (pkg.optionalDependencies) {
    delete pkg.optionalDependencies[brigadierPackageName]
  }
  if (pkg.dependencies) {
    delete pkg.dependencies[brigadierPackageName]
  } else {
    pkg.dependencies = {}
  }
} else {
  logger.debug(`no existing package.json found at ${packageJSONPath}`)
  logger.debug("creating a minimal package.json")
  pkg = {
    private: true,
    dependencies: {}
  }
}
pkg.dependencies[brigadierPackageName] = brigadierSrcPath
logger.debug(`writing package.json to ${packageJSONPath}`)
fs.writeFileSync(packageJSONPath, JSON.stringify(pkg))

// Install dependencies
logger.debug("installing dependencies")
try {
  execFileSync("yarn", ["install"], { cwd: configFilesPath })
} catch(e) {
  throw new Error(`error executing yarn install:\n\n${e.output}`)
}

// Experimental TypeScript support...
if (fs.existsSync(path.join(configFilesPath, "brigade.ts"))) {
  logger.debug("compiling brigade.ts")
  try {
    execFileSync("tsc", ["--target", "ES6", "--module", "commonjs", "brigade.ts"], { cwd: configFilesPath })
  } catch(e) {
    throw new Error(`error compiling brigade.ts:\n\n${e.output}`)
  }
}
