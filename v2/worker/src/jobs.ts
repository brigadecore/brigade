import * as https from "https"

// For some reason, EventSource NEEDS to be required this way.
const EventSource = require("eventsource") // eslint-disable-line @typescript-eslint/no-var-requires

import axios from "axios"

import { Event } from "../../brigadier/src/events"
import { Logger } from "../../brigadier/src/logger"
import { Job as BrigadierJob } from "../../brigadier/src"

import { logger } from "./brigadier"

export class Job extends BrigadierJob {
  logger: Logger

  constructor(name: string, image: string, event: Event) {
    super(name, image, event)
    this.logger = logger.child({ job: name })
  }

  async run(): Promise<void> {
    this.logger.info(`Creating job ${this.name}`)
    try {
      const response = await axios({
        httpsAgent: new https.Agent(
          {
            rejectUnauthorized: false
          }
        ),
        method: "post",
        url: `${this.event.worker.apiAddress}/v2/events/${this.event.id}/worker/jobs`,
        headers: {
          Authorization: `Bearer ${this.event.worker.apiToken}`
        },
        data: {
          apiVersion: "brigade.sh/v2",
          kind: "Job",
          name: this.name,
          spec: {
            primaryContainer: this.primaryContainer,
            sidecarContainers: this.sidecarContainers,
            timeoutSeconds: this.timeout,
            host: this.host
          }
        },
      })
      if (response.status != 201) {
        throw new Error(`Received ${response.status} from the API server`)
      }
    }
    catch(e) {
      throw new Error(`Error creating job "${this.name}": ${e.message}`)
    }
    return this.wait()
  }

  private async wait(): Promise<void> {
    return new Promise<void>((resolve, reject) => {
      const eventSource = new EventSource(
        `${this.event.worker.apiAddress}/v2/events/${this.event.id}/worker/jobs/${this.name}/status?watch=true&sse=true`, 
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
      eventSource.addEventListener("message", (event: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
        let status: any // eslint-disable-line @typescript-eslint/no-explicit-any
        try {
          status = JSON.parse(event.data)
        } catch(e) {
          eventSource.close() 
          reject(new Error(`Error parsing job "${this.name}" status: ${e.message}`))
        }
        this.logger.debug(`Current job phase is ${status.phase}`)
        switch (status.phase) {
        case "ABORTED":
          eventSource.close()
          reject(new Error(`Job "${this.name}" was aborted`))
          break
        case "CANCELED":
          eventSource.close()
          reject(new Error(`Job "${this.name}" was canceled before starting`))
          break
        case "FAILED":
          eventSource.close()
          reject(new Error(`Job "${this.name}" failed`))
          break
        case "SCHEDULING_FAILED":
          eventSource.close()
          reject(new Error(`Job "${this.name}" scheduling failed`))
          break
        case "SUCCEEDED":
          eventSource.close()
          resolve()
          break
        case "TIMED_OUT":
          eventSource.close()
          reject(new Error(`Job "${this.name}" timed out`))
          break
        }
        // For all other phases there's nothing to do. Keep waiting.
      })
      eventSource.addEventListener("error", (e: any) => { // eslint-disable-line @typescript-eslint/no-explicit-any
        if (e.status) { // If the error has an HTTP status code associated with it...
          eventSource.close()
          reject(new Error(`Received ${e.status} from the API server when attempting to open job "${this.name}" status stream`))
        } else if (eventSource.readyState == EventSource.CONNECTING) {
          // We lost the connection and we're reconnecting... nbd
          this.logger.debug("Reconnecting to status stream")
        } else if (eventSource.readyState == EventSource.CLOSED) {
          // We disconnected for some unknown reason... and presumably exhausted
          // attempts to reconnect
          reject(new Error(`Error receiving job "${this.name}" status stream: ${e.message}`))
        }
      })
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
