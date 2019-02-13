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
 *
 * Also, the Brigade script must be written to `brigade.js`.
 */

// Seems to be a bug in typedocs that requires this empty comment.
/** */

import * as fs from "fs";
import * as process from "process";
import * as ulid from "ulid";

import * as events from "@azure/brigadier/out/events";
import { App } from "./app";
import { ContextLogger, LogLevel } from "@azure/brigadier/out/logger";

import { options } from "./k8s";

// This is a side-effect import.
import "./brigade";

const logLevel = LogLevel[process.env.BRIGADE_LOG_LEVEL || "LOG"];
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
  if (!options.serviceAccount.match(`${process.env.BRIGADE_SERVICE_ACCOUNT_REGEX}`)) {
      logger.log(`Service Account ${options.serviceAccount} does not match regex /${process.env.BRIGADE_SERVICE_ACCOUNT_REGEX}/`);
      process.exit(1);
  }
}

// Run the app.
new App(projectID, projectNamespace).run(e);
