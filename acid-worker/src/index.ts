/**
 * The Acid Worker is responsible for executing `acid.js` files.
 *
 * When the Acid Worker starts, it will look for the following environment
 * variables, which it will use to generate AcidEvent and Project configuration:
 *
 * - `ACID_EVENT_TYPE`: The event type, such as `push`, `pull_request`
 * - `ACID_EVENT_PROVIDER`: The name of the event provider, such as `github` or `dockerhub`
 * - `ACID_PROJECT_ID`: The project ID. This is used to load the Project
 *   object from configuration.
 * - `ACID_COMMIT`: The VCS commit ID (e.g. the Git commit)
 * - `ACID_PAYLOAD`: The payload from the original event trigger.
 * - `ACID_PROJECT_NAMESPACE`: For Kubernetes, this is the Kubernetes namespace in
 *   which new jobs should be created. The Acid worker must have write access to
 *   this namespace.
 *
 * Also, the Acid script must be written to `acid.js`.
 */

// Seems to be a bug in typedocs that requires this empty comment.
/** */

import * as fs from 'fs'
import * as process from "process"

import * as events from "./events"
import * as k8s from "./k8s"
import * as libacid from './libacid'
import {App} from "./app"

// This is a side-effect import.
import "./acid"

let projectID: string = process.env.ACID_PROJECT_ID
let projectNamespace: string = process.env.ACID_PROJECT_NAMESPACE
let e: events.AcidEvent = {
    type: process.env.ACID_EVENT_TYPE || "ping",
    provider: process.env.ACID_EVENT_PROVIDER || "unknown",
    commit: process.env.ACID_COMMIT || "master",
}

try {
  e.payload = fs.readFileSync("/etc/acid/payload", "utf8")
} catch (e) {
  console.log("no payload loaded")
}

// Run the app.
(new App(projectID, projectNamespace)).run(e)
