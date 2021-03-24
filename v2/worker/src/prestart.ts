import * as fs from "fs"
import { execFileSync } from "child_process"
import * as path from "path"

import { Event, logger } from "@brigadecore/brigadier-polyfill"

logger.info(`brigade-worker version: ${process.env.WORKER_VERSION}`)

const event: Event = require("/var/event/event.json") // eslint-disable-line @typescript-eslint/no-var-requires

logger.level = (event.worker.logLevel || "info").toLowerCase()
const debug = logger.level == "debug"

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

const packageJSONPath = path.join(configFilesPath, "package.json")
if (fs.existsSync(packageJSONPath)) {
  logger.debug(`found a package.json at ${packageJSONPath}`)
  // Install dependencies
  // If we find package-lock.json, we use npm. Otherwise, we use yarn.
  const packageLockJSONPath = path.join(configFilesPath, "package-lock.json")
  if (fs.existsSync(packageLockJSONPath)) {
    logger.debug("installing dependencies using npm")
    try {
      execFileSync("npm", ["install", "--prod"], {
        cwd: configFilesPath,
        stdio: debug ? "inherit" : undefined
      })
    } catch(e) {
      throw new Error(`error executing npm install:\n\n${e.output}`)
    }
  } else {
    logger.debug("installing dependencies using yarn")
    try {
      execFileSync("yarn", ["install", "--prod"], {
        cwd: configFilesPath,
        stdio: debug ? "inherit" : undefined
      })
    } catch(e) {
      throw new Error(`error executing yarn install:\n\n${e.output}`)
    }
  }  
}

const moduleNamespace = "@brigadecore"
const moduleNamespacePath = path.join(configFilesPath, "node_modules", moduleNamespace)
if (!fs.existsSync(moduleNamespacePath)) {
  logger.debug(`path ${moduleNamespacePath} does not exist; creating it`)
  fs.mkdirSync(moduleNamespacePath, { recursive: true })
}

const moduleName = "brigadier"
const modulePath = path.join(moduleNamespacePath, moduleName)
if (fs.existsSync(modulePath)) {
  logger.debug(`path ${modulePath} exists; deleting it`)
  fs.rmSync(modulePath, { recursive: true, force: true })
}

const modulePolyfillPath = "/var/brigade-worker/brigadier-polyfill"
logger.debug(`polyfilling ${moduleNamespace}/${moduleName} with ${modulePolyfillPath}`)
fs.symlinkSync(modulePolyfillPath, modulePath)

// Experimental TypeScript support...
if (fs.existsSync(path.join(configFilesPath, "brigade.ts"))) {
  logger.debug("compiling brigade.ts")
  try {
    execFileSync("tsc", ["--target", "ES6", "--module", "commonjs", "brigade.ts"], {
      cwd: configFilesPath,
      stdio: debug ? "inherit" : undefined
    })
  } catch(e) {
    throw new Error(`error compiling brigade.ts:\n\n${e.output}`)
  }
}

try {
  execFileSync("node", ["brigade.js"], {
    cwd: configFilesPath,
    stdio: "inherit"
  })
} catch(e) {
  throw new Error(`error executing brigade.js:\n\n${e.output}`)
}
