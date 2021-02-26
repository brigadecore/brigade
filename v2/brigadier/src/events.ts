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
  /** The gateway which created the event. */
  source: string
  /** The type of event. Values and meanings are source-specific. */
  type: string
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
  public on(eventSource: string, eventType: string, eventHandler: EventHandler): this {
    this.handlers[`${eventSource}:${eventType}`] = eventHandler
    return this
  }
}

/** Contains event handler registrations for a script. */
export const events = new EventRegistry()
