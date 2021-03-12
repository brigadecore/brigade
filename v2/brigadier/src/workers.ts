/**
 * Represents the Brigade worker within which an event handler runs.
 */
export interface Worker {
  /** The address of the Brigade API server. */
  apiAddress: string
  /**
   * A token which can be used to authenticate to the API server.
   * The token is specific to the current event and allows you to create
   * jobs for that event. It has no other permissions.
   */
  apiToken: string
  /**
   * The directory where the worker stores configuration files,
   * including event handler code files such as brigade.js and brigade.json.
   */
  configFilesDirectory: string
  /**
   * The default values to use for any configuration files that are not present.
   */
  defaultConfigFiles: { [key: string]: string }
  /**
   * The desired granularity of worker logs. Worker logs are distinct from job
   * logs - the containers in a job will emit logs according to their own
   * configuration.
   */
  logLevel?: string
}
