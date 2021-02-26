/**
 * A Brigade project.
 */
export interface Project {
  /** The unique identifier of the project. */
  id: string
  /** A map of secrets defined in the Brigade project. */
  secrets: { [key: string]: string }
}
