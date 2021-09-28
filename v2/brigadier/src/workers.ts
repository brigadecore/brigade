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
   * including event handler code files such as brigade.js and package.json.
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
  /**
   * If applicable, contains git-specific Worker details.
   */
  git?: GitConfig
}

/**
 * Represents git-specific Worker details.
 */
export interface GitConfig {
  /**
   * Specifies the remote repository where, if applicable, the Worker will have
   * obtained source code from.
   */
  cloneURL: string
  /**
   * Specifies the exact commit where, if applicable, the Worker will have
   * obtained source code from.
   */
  commit?: string
  /**
   * Specifies a symbolic reference to the commit where, if applicable, the
   * Worker will have obtained source code from.
   *
   * It is sometimes useful for Brigade scripts (e.g. brigade.js or brigade.ts)
   * to examine the value of this field to determine, for instance, the name of
   * a branch or tag related to the Event the Worker is handling.
   */
  ref?: string
}
