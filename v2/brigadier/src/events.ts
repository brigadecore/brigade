import { Worker } from "./workers"
import { Project } from "./projects"

/**
 * An event received by Brigade that resulted in triggering an event handler.
 */
export interface Event {
  /** A unique identifier for the event. */
  id: string
  /** The project that registered the handler being called for the event. */
  project: Project
  /** The unique identifier of the gateway which created the event. */
  source: string
  /** The type of event. Values and meanings are source-specific. */
  type: string
  /** Event's type disambiguators. */
  qualifiers?: { [key: string]: string }
  /** Supplementary Event details. */
  labels?: { [key: string]: string }
  /** A short title for the event, suitable for display in space-limited UI such as lists. */
  shortTitle?: string
  /** A detailed title for the event. */
  longTitle?: string
  /** The content of the event. This is source- and type-specific. */
  payload?: string
  /** The Brigade worker assigned to handle the event. */
  worker: Worker
}

/** The type of the procedure to handle an Event. */
export type EventHandler = (event: Event) => void

/**
 * Contains event handler registrations for a script.
 *
 * Access the registry through the global `events` object.
 */
export class EventRegistry {
  protected handlers: { [key: string]: EventHandler } = {}

  /**
   * Registers a handler to be run when a particular kind of event occurs.
   * When Brigade receives a matching event, it will run the specified
   * handler.
   *
   * Possible event sources depend on the gateways configured in your Brigade
   * system, and possible event types depend on the gateway. See the Brigade
   * and gateway documentation for details.
   *
   * @param eventSource The event source (gateway) to register the handler for
   * @param eventType The event type to register the handler for
   * @param eventHandler The handler to run when an event with the given source and
   * type occurs
   */
  public on(
    eventSource: string,
    eventType: string,
    eventHandler: EventHandler
  ): this {
    this.handlers[`${eventSource}:${eventType}`] = eventHandler
    return this
  }

  public process(): void {
    let event: Event
    if (process.env.BRIGADE_EVENT_FILE) {
      console.log(
        `Loading dummy event from file ${process.env.BRIGADE_EVENT_FILE}`
      )
      event = require(process.env.BRIGADE_EVENT_FILE)
    } else {
      console.log("No dummy event file provided")
      console.log("Generating a dummy event")
      let source: string
      let type: string
      if (process.env.BRIGADE_EVENT) {
        let eventTokens: string[] = []
        const match = process.env.BRIGADE_EVENT.match(/^(.+?):(.+)$/)
        if (match) {
          eventTokens = Array.from(match)
        }
        if (eventTokens.length != 3) {
          throw new Error(
            `${process.env.BRIGADE_EVENT} is not a valid event of the form <source>:<type>`
          )
        }
        source = eventTokens[1]
        type = eventTokens[2]
        console.log(
          `Using specified dummy event with source "${source}" and type "${type}"`
        )
      } else {
        console.log("No dummy event type provided")
        source = "brigade.sh/cli"
        type = "exec"
        console.log(
          `Using default dummy event with source "${source}" and type "${type}"`
        )
      }
      event = {
        id: this.newUUID(),
        source: source,
        type: type,
        project: {
          id: this.newUUID(),
          secrets: {}
        },
        worker: {
          apiAddress: "https://brigade2.example.com",
          apiToken: this.newUUID(),
          configFilesDirectory: ".brigade",
          defaultConfigFiles: {}
        }
      }
    }
    console.log("Processing the following dummy event:")
    console.log(event)
    this.fire(event)
  }

  protected fire(event: Event): string | void {
    let handlerFn = this.handlers[`${event.source}:${event.type}`]
    if (!handlerFn) {
      handlerFn = this.handlers[`${event.source}:*`]
    }
    if (handlerFn) {
      return handlerFn(event)
    }
  }

  private newUUID(): string {
    return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(
      /[xy]/g,
      function (c) {
        const r = (Math.random() * 16) | 0,
          v = c == "x" ? r : (r & 0x3) | 0x8
        return v.toString(16)
      }
    )
  }
}

/** Contains event handler registrations for a script. */
export const events = new EventRegistry()
