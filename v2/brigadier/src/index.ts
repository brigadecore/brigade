import { Event, EventRegistry } from "./events"
import { Group } from "./groups"
import { Job, Container, JobHost } from "./jobs"
import { logger } from "./logger"

// events is the main event registry
export const events = new EventRegistry()

export { Event, Group, Job, Container, JobHost, logger }
