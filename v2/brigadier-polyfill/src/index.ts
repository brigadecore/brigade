// This module provides the worker with a polyfill (drop-in replacement) for the
// REAL Brigadier module. This drop-in replacement seamlessly replaces certain
// Brigadier behaviors with more sophisticated alternatives. For instance, the
// no-op job.run() function is replaced with one that communicates with the
// Brigade API server.

// Export these things directly from Brigadier
export {
  Container,
  ConcurrentGroup,
  Event,
  JobHost,
  SerialGroup
} from "@brigadecore/brigadier"

// These are custom implementations of resources ordinarily found in Brigadier
export { events } from "./events"
export { Job } from "./jobs"
export { logger } from "./logger"
