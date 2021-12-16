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
      const jobsClient = new core.JobsClient(
        this.event.worker.apiAddress,
        this.event.worker.apiToken,
        {allowInsecureConnections: true},
      )

      const statusStream = jobsClient.watchStatus(this.event.id, this.name)
      statusStream.onData((status: core.JobStatus) => {
        this.logger.debug(`Current job phase is ${status.phase}`)
        if (!this.fallible) {
          switch (status.phase) {
          case core.JobPhase.Aborted:
            reject(new Error(`Job "${this.name}" was aborted`))
            break
          case core.JobPhase.Canceled:
            reject(new Error(`Job "${this.name}" was canceled before starting`))
            break
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
        }
      })
      statusStream.onReconnecting(() => {
        this.logger.warn("status stream connecting")
      })
      statusStream.onClosed(() => {
        const msg = "status stream closed"
        if (this.fallible) {
          this.logger.warn(msg)
          resolve()
        } else {
          reject(new Error(msg)) 
        }
      })
      statusStream.onError((e: Error) => {
        const msg = `Error watching status for job "${this.name}": ${e.message}`
        if (this.fallible) {
          this.logger.warn(msg)
          resolve()
        } else {
          reject(new Error(msg))
        }
      })
      statusStream.onDone(() => {
        resolve()
      })
    })
  }

  async logs(): Promise<string> {
    return new Promise<string>((resolve, reject) => {
      const logsClient = new core.LogsClient(
        this.event.worker.apiAddress,
        this.event.worker.apiToken,
        {allowInsecureConnections: true},
      )

      const logsStream = logsClient.stream(
        this.event.id,
        {job: this.name},
        {follow: false},
      )
      let logs = ""
      logsStream.onData((logEntry: core.LogEntry) => {
        if (logs != "") {
          logs += "\n"
        }
        if (logEntry.time) {
          logs += logEntry.time + ": "
        }
        logs += logEntry.message
      })
      logsStream.onReconnecting(() => {
        this.logger.warn("log stream connecting")
      })
      logsStream.onClosed(() => {
        reject("log stream closed")
      })
      logsStream.onError((e: Error) => {
        reject(new Error(`Error retrieving logs for job "${this.name}": ${e.message}`))
      })
      logsStream.onDone(() => {
        resolve(logs)
      })
    })
  }
}
