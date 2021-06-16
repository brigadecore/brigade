// For some reason, EventSource NEEDS to be required this way.
const EventSource = require("eventsource") // eslint-disable-line @typescript-eslint/no-var-requires

import { Logger } from "winston" 

import { Event, Job as BrigadierJob } from "@brigadecore/brigadier"

import { core } from "@brigadecore/brigade-sdk"

import { logger } from "./logger"

export class Job extends BrigadierJob {
  logger: Logger

  constructor(name: string, image: string, event: Event) {
    super(name, image, event)
    this.logger = logger.child({ job: name })
  }

  async run(): Promise<void> {
    this.logger.info(`Creating job ${this.name}`)
    try {
      const jobsClient = new core.JobsClient(
        this.event.worker.apiAddress,
        this.event.worker.apiToken,
        {allowInsecureConnections: true},
      )

      const sdkJob: core.Job = {
        name: this.name,
        spec: {
          primaryContainer: this.primaryContainer,
          sidecarContainers: this.sidecarContainers,
          timeoutDuration: this.timeoutSeconds + "s",
          host: this.host
        }
      }
      await jobsClient.create(this.event.id, sdkJob)
    }
    catch(e) {
      throw new Error(`Error creating job "${this.name}": ${e.message}`)
    }
    return this.wait()
  }

  private async wait(): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      try {
        const jobsClient = new core.JobsClient(
          this.event.worker.apiAddress,
          this.event.worker.apiToken,
          {allowInsecureConnections: true},
        )

        const statusStream = jobsClient.watchStatus(this.event.id, this.name)
        statusStream.onData((status: core.JobStatus) => {
          this.logger.debug(`Current job phase is ${status.phase}`)
          switch (status.phase) {
          case core.JobPhase.Aborted:
            reject(new Error(`Job "${this.name}" was aborted`))
            break
          // TODO: uncomment once SDK has core.JobPhase.Canceled
          // case core.JobPhase.Canceled:
          //   reject(new Error(`Job "${this.name}" was canceled before starting`))
          //   break
          case core.JobPhase.Failed:
            reject(new Error(`Job "${this.name}" failed`))
            break
          case core.JobPhase.SchedulingFailed:
            reject(new Error(`Job "${this.name}" scheduling failed`))
            break
          case core.JobPhase.Succeeded:
            resolve()
            break
          case core.JobPhase.TimedOut:
            reject(new Error(`Job "${this.name}" timed out`))
            break
          }
        })
        statusStream.onReconnecting(() => {
          console.log("status stream connecting")
        })
        statusStream.onClosed(() => {
          reject("status stream closed")
        })
        statusStream.onError((e: Error) => {
          reject(e)
        })
        statusStream.onDone(() => {
          resolve()
        })
      }
      catch(e) {
        throw new Error(`Error watching status for job "${this.name}": ${e.message}`)
      }
    })
  }

  async logs(): Promise<string> {
    return new Promise<string>((resolve, reject) => {
      const eventSource = new EventSource(
        `${this.event.worker.apiAddress}/v2/events/${this.event.id}/logs?job=${this.name}&sse=true`, 
        {
          https: {
            // TODO: Get our hands on the API server's CA to validate the cert
            rejectUnauthorized: false
          },
          headers: {
            "Authorization": `Bearer ${this.event.worker.apiToken}`
          }
        }
      )
      let logs = ""
      eventSource.addEventListener("message", (event: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
        let logEntry: any // eslint-disable-line @typescript-eslint/no-explicit-any
        try {
          logEntry = JSON.parse(event.data)
        } catch(e) {
          eventSource.close() 
          reject(new Error(`Error parsing log entry for job "${this.name}": ${e.message}`))
        }
        if (logs != "") {
          logs += "\n"
        }
        logs += logEntry.message
      })
      eventSource.addEventListener("error", (e: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
        if (e.status) { // If the error has an HTTP status code associated with it...
          eventSource.close()
          reject(new Error(`Received ${e.status} from the API server when attempting to open job "${this.name}" log stream`))
        } else if (eventSource.readyState == EventSource.CONNECTING) {
          // We lost the connection and we're reconnecting... nbd
          this.logger.debug("Reconnecting to log stream")
        } else if (eventSource.readyState == EventSource.CLOSED) {
          // We disconnected for some unknown reason... and presumably exhausted
          // attempts to reconnect
          reject(new Error(`Encountered unknown error receiving job "${this.name}" log stream`))
        }
      })
      eventSource.addEventListener("done", () => {
        eventSource.close()
        resolve(logs)
      })
    })
  }

}
