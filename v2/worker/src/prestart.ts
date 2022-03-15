import * as fs from "fs"
import { execFileSync, ExecFileSyncOptions } from "child_process"
import * as path from "path"

import { Event, logger } from "@brigadecore/brigadier-polyfill"

const nodePath = "/nodejs/bin/node"
const npmPath = "/var/brigade-worker/worker/node_modules/.bin/npm"
const yarnPath = "/var/brigade-worker/worker/node_modules/.bin/yarn"
const tscPath = "/var/brigade-worker/worker/node_modules/.bin/tsc"

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

// Figure out whether we should use npm or yarn as the package manager. If we
// find yarn.lock, we use yarn. Otherwise, we use npm.
const yarnLockPath = path.join(configFilesPath, "yarn.lock")
let useYarn = false
if (fs.existsSync(yarnLockPath)) {
  logger.debug("using yarn as the package manager")
  useYarn = true
} else {
  logger.debug("using npm as the package manager")
}

// If we can find a package.json, load it into an object
let packageJSON: any // eslint-disable-line @typescript-eslint/no-explicit-any
const packageJSONPath = path.join(configFilesPath, "package.json")
if (fs.existsSync(packageJSONPath)) {
  logger.debug(`found a package.json at ${packageJSONPath}`)
  packageJSON = require(packageJSONPath)
}

// prepExecOpts-- include more or less output in the worker's own logs,
// depending on whether we are debugging or not.
const prepExecOpts: ExecFileSyncOptions = {
  cwd: configFilesPath,
  stdio: debug ? "inherit" : undefined
}

const moduleNamespace = "@brigadecore"
const moduleName = "brigadier"

// Install dependencies, if any
if (packageJSON && packageJSON.dependencies) {
  // Remove @brigadecore/brigade from the dependencies since we're going to
  // polyfill it anyway.
  if (packageJSON.dependencies[`${moduleNamespace}/${moduleName}`]) {
    logger.debug(`deleting ${moduleNamespace}/${moduleName} from package.json`)
    delete packageJSON.dependencies[`${moduleNamespace}/${moduleName}`]
    fs.writeFileSync(packageJSONPath, JSON.stringify(packageJSON))
  }
  if (Object.keys(packageJSON.dependencies).length === 0) {
    logger.debug("no dependencies -- bypassing dependency resolution")
  } else {
    useYarn ? yarnInstall() : npmInstall()
  }
}

// Add/replace @brigadecore/brigadier with the worker's brigadier polyfill
const moduleNamespacePath = path.join(configFilesPath, "node_modules", moduleNamespace)
if (!fs.existsSync(moduleNamespacePath)) {
  logger.debug(`path ${moduleNamespacePath} does not exist; creating it`)
  fs.mkdirSync(moduleNamespacePath, { recursive: true })
}
const modulePath = path.join(moduleNamespacePath, moduleName)
if (fs.existsSync(modulePath)) {
  logger.debug(`path ${modulePath} exists; deleting it`)
  fs.rmSync(modulePath, { recursive: true, force: true })
}
const modulePolyfillPath = "/var/brigade-worker/brigadier-polyfill"
logger.debug(`polyfilling ${moduleNamespace}/${moduleName} with ${modulePolyfillPath}`)
fs.symlinkSync(modulePolyfillPath, modulePath)

// Build, if applicable
if (packageJSON?.scripts?.build) {
  useYarn ? yarnBuild() : npmBuild()
} else if (fs.existsSync(path.join(configFilesPath, "tsconfig.json"))) {
  compileWithTSCConfig()
} else if (fs.existsSync(path.join(configFilesPath, "brigade.ts"))) {
  defaultCompile()
} else {
  logger.debug("found nothing to compile")
}

// runExecOpts-- unconditionally include output in the worker's own logs.
const runExecOpts: ExecFileSyncOptions = {
  cwd: configFilesPath,
  stdio: "inherit"
}

// Now run
if (packageJSON?.scripts?.run) {
  useYarn ? yarnRun() : npmRun()
} else if (fs.existsSync(path.join(configFilesPath, "brigade.js"))) {
  nodeRun()
} else {
  throw new Error("found nothing to run")
}

function yarnInstall(): void {
  logger.debug("installing dependencies using yarn")
  try {
    execFileSync(nodePath, [yarnPath, "install", "--prod"], prepExecOpts)
  } catch(e) {
    throw new Error(`error executing yarn install:\n\n${e.output}`)
  }
}

function npmInstall(): void {
  logger.debug("installing dependencies using npm")
  try {
    execFileSync(nodePath, [npmPath, "install", "--prod"], prepExecOpts)
  } catch(e) {
    throw new Error(`error executing npm install:\n\n${e.output}`)
  }
}

function yarnBuild(): void {
  logger.debug("running build script with yarn")
  try {
    execFileSync(nodePath, [yarnPath, "build"], prepExecOpts)
  } catch(e) {
    throw new Error(`error executing yarn build:\n\n${e.output}`)
  } 
}

function npmBuild(): void {
  logger.debug("running build script with npm")
  try {
    execFileSync(nodePath, [npmPath, "run-script", "build"], prepExecOpts)
  } catch(e) {
    throw new Error(`error executing npm run-script build:\n\n${e.output}`)
  }
}

function compileWithTSCConfig() {
  logger.debug("compiling typescript project with configuration from tsconfig.json")
  try {
    execFileSync(nodePath, [tscPath], prepExecOpts)
  } catch(e) {
    throw new Error(`error executing tsc:\n\n${e.output}`)
  }
}

function defaultCompile() {
  logger.debug("compiling brigade.ts with flags --target ES6 --module commonjs --esModuleInterop")
  try {
    execFileSync(nodePath, [tscPath, "--target", "ES6", "--module", "commonjs", "--esModuleInterop", "brigade.ts"], prepExecOpts)
  } catch(e) {
    throw new Error(`error compiling brigade.ts:\n\n${e.output}`)
  }
}

function yarnRun(): void {
  logger.debug("running script with yarn")
  try {
    execFileSync(nodePath, [yarnPath, "run", "run"], runExecOpts)
  } catch(e) {
    throw new Error(`error executing yarn run run:\n\n${e.output}`)
  }
}

function npmRun(): void {
  logger.debug("running script with npm")
  try {
    execFileSync(nodePath, [npmPath, "run-script", "run"], runExecOpts)
  } catch(e) {
    throw new Error(`error executing npm run-script run:\n\n${e.output}`)
  }
}

function nodeRun(): void {
  logger.debug("running node brigade.js")
  try {
    execFileSync(nodePath, ["brigade.js"], runExecOpts)
  } catch(e) {
    throw new Error(`error executing node brigade.js:\n\n${e.output}`)
  }
}
