/**
 * The Brigade Worker is responsible for executing `brigade.js` files.
 *
 * When the Brigade Worker starts, it will look for the following environment
 * variables, which it will use to generate BrigadeEvent and Project configuration:
 *
 * - `BRIGADE_EVENT_TYPE`: The event type, such as `push`, `pull_request`
 * - `BRIGADE_EVENT_PROVIDER`: The name of the event provider, such as `github` or `dockerhub`
 * - `BRIGADE_PROJECT_ID`: The project ID. This is used to load the Project
 *   object from configuration.
 * - `BRIGADE_REMOTE_URL`: The URL from which to obtain source code to be built.
 *   This is optional. If left unset by the controller, the worker will fall
 *   back to a project-level URL.
 * - `BRIGADE_COMMIT_ID`: The VCS commit ID (e.g. the Git commit)
 * - `BRIGADE_COMMIT_REF`: The VCS full reference, defaults to
 *   `refs/heads/master`
 * - `BRIGADE_PAYLOAD`: The payload from the original event trigger.
 * - `BRIGADE_PROJECT_NAMESPACE`: For Kubernetes, this is the Kubernetes namespace in
 *   which new jobs should be created. The Brigade worker must have write access to
 *   this namespace.
 * - `BRIGADE_BUILD_ID`: The ULID for the build. This is unique.
 * - `BRIGADE_BUILD_NAME`: This is actually the ID of the worker.
 * - `BRIGADE_SERVICE_ACCOUNT`: The service account to use.
 * - `BRIGADE_DEFAULT_BUILD_STORAGE_CLASS`: The Kubernetes StorageClass to use
 *   for shared build storage if none is specified in project configuration.
 * - `BRIGADE_DEFAULT_CACHE_STORAGE_CLASS`: The Kubernetes StorageClass to use
 *   for caching jobs if none is specified in project configuration.
 *
 * Also, the Brigade script must be written to `brigade.js`.
 */

// Seems to be a bug in typedocs that requires this empty comment.
/** */

import * as fs from "fs";
import * as moduleAlias from "module-alias";
import * as path from "path";
import * as process from "process";
import * as ulid from "ulid";

import * as events from "@brigadecore/brigadier/out/events";
import { App } from "./app";
import { ContextLogger, LogLevel } from "@brigadecore/brigadier/out/logger";

import { options } from "./k8s";

// Script locations in order of precedence.
const scripts = [
  // manual override for debugging
  process.env.BRIGADE_SCRIPT,

  // data mounted from event secret (e.g. brig run)
  "/etc/brigade/script",

  // checked out in repo
  "/vcs/brigade.js",

  // data mounted from project.DefaultScript
  "/etc/brigade-project/defaultScript",

  // mounted configmap named in brigade.sh/project.DefaultScriptName
  "/etc/brigade-default-script/brigade.js"
];

function findScript() {
  for (let src of scripts) {
    if (fs.existsSync(src) && fs.readFileSync(src, "utf8") != "") {
      return src;
    }
  }
}

// Search for the Brigade script and, if found, execute it.
const script = findScript();
if (script) {
  // Install aliases for common ways of referring to Brigade/Brigadier.
  moduleAlias.addAliases({
    "brigade": __dirname + "/brigadier",
    "brigadier": __dirname + "/brigadier",
    "@brigadecore/brigadier": __dirname + "/brigadier",
  });

  // Add the current module resolution paths to module-alias, so the node_modules
  // that prestart.js adds to will be resolvable from the Brigade script and any
  // local dependencies.
  module.paths.forEach(moduleAlias.addPath);

  const realScriptPath = fs.realpathSync(script);
  // NOTE: `as any` is needed because @types/module-alias is at 2.0.0, while
  //       module-alias is now at 2.2.0.
  (moduleAlias as any).addAlias(".", (fromPath: string) => {
    // A custom handler for local dependencies to handle cases where the entry
    // script is outside `/vcs`.

    // For entry scripts outside /vcs only, rewrite dot-slash-prefixed requires
    // to be rooted at `/vcs`.
    if (!fromPath.startsWith("/vcs") && fromPath === realScriptPath) {
      return "/vcs";
    }

    // For all other dot-slash-prefixed requires, resolve as usual.
    // NOTE: module-alias will not allow us to just return "." here, because
    // it uses path.join under the hood, which collapses "./foo" down to just
    // "foo", for which the module resolution semantics are different.  So,
    // return the directory of the requiring module, which gives the same result
    // as ".".
    return path.dirname(fromPath);
  });

  moduleAlias();
  require(script);
}

// Log level may come in as lowercased 'log', 'info', etc., if run by the brig cli
const logLevel = LogLevel[process.env.BRIGADE_LOG_LEVEL.toUpperCase() || "LOG"];
const logger = new ContextLogger([], logLevel);

const version = require("../package.json").version;
logger.log(`brigade-worker version: ${version}`);

const requiredEnvVar = (name: string): string => {
  if (!process.env[name]) {
    logger.log(`Missing required env ${name}`);
    process.exit(1);
  }
  return process.env[name];
};

const projectID: string = requiredEnvVar("BRIGADE_PROJECT_ID");
const projectNamespace: string = requiredEnvVar("BRIGADE_PROJECT_NAMESPACE");
const defaultULID = ulid().toLocaleLowerCase();
let e: events.BrigadeEvent = {
  buildID: process.env.BRIGADE_BUILD_ID || defaultULID,
  workerID: process.env.BRIGADE_BUILD_NAME || `unknown-${defaultULID}`,
  type: process.env.BRIGADE_EVENT_TYPE || "ping",
  provider: process.env.BRIGADE_EVENT_PROVIDER || "unknown",
  cloneURL: process.env.BRIGADE_REMOTE_URL,
  revision: {
    commit: process.env.BRIGADE_COMMIT_ID,
    ref: process.env.BRIGADE_COMMIT_REF
  },
  logLevel: logLevel
};

try {
  e.payload = fs.readFileSync("/etc/brigade/payload", "utf8");
} catch (e) {
  logger.log("no payload loaded");
}

if (process.env.BRIGADE_SERVICE_ACCOUNT) {
  options.serviceAccount = process.env.BRIGADE_SERVICE_ACCOUNT;
}

if (process.env.BRIGADE_SERVICE_ACCOUNT_REGEX) {
  let regex = RegExp(`${process.env.BRIGADE_SERVICE_ACCOUNT_REGEX}`);
  if (!regex.test(options.serviceAccount)) {
      logger.log(`Service Account ${options.serviceAccount} does not match regex ${process.env.BRIGADE_SERVICE_ACCOUNT_REGEX}`);
      process.exit(1);
  }
}

if (process.env.BRIGADE_DEFAULT_BUILD_STORAGE_CLASS) {
  options.defaultBuildStorageClass = process.env.BRIGADE_DEFAULT_BUILD_STORAGE_CLASS
}
if (process.env.BRIGADE_DEFAULT_CACHE_STORAGE_CLASS) {
  options.defaultCacheStorageClass = process.env.BRIGADE_DEFAULT_CACHE_STORAGE_CLASS
}

// Run the app.
new App(projectID, projectNamespace).run(e);
