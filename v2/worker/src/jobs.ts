import * as https from "https"
import * as http2 from "http2"

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
        method: "put",
        url: `${this.event.worker.apiAddress}/v2/events/${this.event.id}/worker/jobs/${this.name}`,
        headers: {
          Authorization: `Bearer ${this.event.worker.apiToken}`
        },
        data: {
          apiVersion: "brigade.sh/v2",
          kind: "Job",
          spec: {
            primaryContainer: this.primaryContainer,
            sidecarContainers: this.sidecarContainers,
            timeoutSeconds: this.timeout,
            host: this.host
          }
        },
      })
      if (response.status != 201) {
        console.log(response.data)
        throw new Error(response.data)
      }
    }
    catch(err) {
      // Wrap the original error to give clear context.
      throw new Error(`job ${this.name}: ${err}`)
    }
    return this.wait()
  }

  private async wait(): Promise<void> {
    return new Promise((resolve, reject) => {
      let abortMonitor = false
      let req: http2.ClientHttp2Stream
      
      const startMonitorReq = () => {
        const client = http2.connect(
          this.event.worker.apiAddress,
          {
            // TODO: Get our hands on the API server's CA to validate the cert
            rejectUnauthorized: false,
          }
        )
        client.on("error", (err) => console.error(err))
        req = client.request({
          ":path": `/v2/events/${this.event.id}/worker/jobs/${this.name}/status?watch=true`,
          "Authorization": `Bearer ${this.event.worker.apiToken}`
        })
        req.setEncoding("utf8")

        req.on("response", (response) => {
          const status = response[":status"]
          if (status != 200) {
            reject(new Error(`Received ${status} when attempting to stream job status`))
            abortMonitor = true
            req.destroy()
          }
        })

        req.on("data", (data: string) => {
          try {
            const status = JSON.parse(data)
            this.logger.debug(`Job phase is ${status.phase}`)
            switch (status.phase) {
            // TODO: Do we still use this phase???
            case "ABORTED":
              reject(new Error("Job was aborted"))
              abortMonitor = true
              req.destroy()
              break
            case "FAILED":
              reject(new Error("Job failed"))
              abortMonitor = true
              req.destroy()
              break
            case "SUCCEEDED":
              resolve()
              abortMonitor = true
              req.destroy()
              break
            case "TIMED_OUT":
              reject(new Error("Job timed out"))
              abortMonitor = true
              req.destroy()
            }
          } catch (e) {
            // Let it stay connected
          } 
        })

        req.on("end", () => {
          client.destroy()
          if (!abortMonitor) {
            // We got disconnected, but apparently not deliberately, so try
            // again.
            this.logger.debug("Had to restart the job monitor")
            startMonitorReq()
          }
        })
      }
      startMonitorReq() // This starts the monitor for the first time.
    })
  }

  async logs(): Promise<string> {
    return new Promise((resolve, reject) => {
      let logs = ""

      const client = http2.connect(
        this.event.worker.apiAddress,
        {
          // TODO: Get our hands on the API server's CA to validate the cert
          rejectUnauthorized: false,
        }
      )
      client.on("error", (err) => console.error(err))
      
      const req = client.request({
        ":path": `/v2/events/${this.event.id}/logs?job=${this.name}`,
        "Authorization": `Bearer ${this.event.worker.apiToken}`
      })
      req.setEncoding("utf8")

      req.on("response", (response) => {
        const status = response[":status"]
        if (status != 200) {
          reject(new Error(`Received ${status} when attempting to stream job logs`))
          req.destroy()
        }
      })

      req.on("data", (data: string) => {
        try {
          const logEntry = JSON.parse(data)
          if (logs != "") {
            logs += "\n"
          }
          logs += logEntry.message
        } catch (e) {
          reject(e)
          req.destroy()
        }
      })

      req.on("end", () => {
        resolve(logs)
        client.destroy()
      })
    })
  }

}
