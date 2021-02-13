import { Worker } from "./workers"
import { Project } from "./projects"

export interface Event {
  id: string
  project: Project
  source: string
  type: string
  shortTitle?: string
  longTitle?: string
  payload?: string
  worker: Worker
}

export type EventHandler = (event: Event) => void

export class EventRegistry {
  protected handlers: { [key: string]: EventHandler } = {}
  
  public on(eventSource: string, eventType: string, eventHandler: EventHandler): this {
    this.handlers[`${eventSource}:${eventType}`] = eventHandler
    return this
  }
}

export const events = new EventRegistry()
