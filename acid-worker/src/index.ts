import * as events from "./events"
import * as process from "process"
import * as k8s from "./k8s"
import * as libacid from './libacid'

// This is a side-effect import.
import "./acid"

let e: events.AcidEvent = {
    type: envOrDefault("ACID_EVENT_TYPE", "ping"),
    provider: envOrDefault("ACID_EVENT_PROVIDER", "unknown"),
    commit: envOrDefault("ACID_COMMIT", "master"),
    // TODO: I think it is safer to read this from a file.
    // TODO: Should we decode this using JSON.parse?
    payload: process.env["ACID_PAYLOAD"]
}

let projectName: string = process.env["ACID_PROJECT_ID"]
let projectNamespace: string = process.env["ACID_PROJECT_NAMESPACE"]

process.on("unhandledRejection", (reason: any, p: Promise<any>) => {
  console.log(`FATAL: ${ reason } (rejection)`)
  process.exit(3)
})

// This traps unhandled 'throw' calls, and is considered safer than
// process.on("unhandledException").
try {
  k8s.loadProject(projectName, projectNamespace).then((p) => {
    libacid.fire(e, p)
  })
} catch (e) {
  console.log(`FATAL: ${ e } (exception)`)
  process.exit(3)
}

function envOrDefault(name: string, defaultValue: string): string {
    let ret = process.env[name]
    if (ret === undefined) {
        return defaultValue;
    }
    return ret;
}
