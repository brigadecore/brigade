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
 * - `BRIGADE_COMMIT`: The VCS commit ID (e.g. the Git commit)
 * - `BRIGADE_PAYLOAD`: The payload from the original event trigger.
 * - `BRIGADE_PROJECT_NAMESPACE`: For Kubernetes, this is the Kubernetes namespace in
 *   which new jobs should be created. The Brigade worker must have write access to
 *   this namespace.
 * - `BRIGADE_BUILD`: The ULID for the build. This is unique.
 * - `BRIGADE_BUILD_NAME`: This is actually the ID of the worker.
 *
 * Also, the Brigade script must be written to `brigade.js`.
 */

// Seems to be a bug in typedocs that requires this empty comment.
/** */

import * as fs from "fs"
import * as process from "process"

import * as ulid from "ulid"

import * as events from "./events"
import * as k8s from "./k8s"
import * as brigadier from './brigadier'
import {App} from "./app"

// This is a side-effect import.
import "./brigade"

const pkg = require('../../package.json')
console.log(`brigade-worker version: ${pkg.version}`)

let projectID: string = process.env.BRIGADE_PROJECT_ID
let projectNamespace: string = process.env.BRIGADE_PROJECT_NAMESPACE
let commit = process.env.BRIGADE_COMMIT || "master"
let defaultULID = ulid()
let e: events.BrigadeEvent = {
    buildID: process.env.BRIGADE_BUILD || defaultULID,
    workerID: process.env.BRIGADE_BUILD_NAME || `unknown-${ defaultULID }`,
    type: process.env.BRIGADE_EVENT_TYPE || "ping",
    provider: process.env.BRIGADE_EVENT_PROVIDER || "unknown",
    commit: commit
}

try {
  e.payload = fs.readFileSync("/etc/brigade/payload", "utf8")
} catch (e) {
  console.log("no payload loaded")
}

// Run the app.
(new App(projectID, projectNamespace)).run(e)
