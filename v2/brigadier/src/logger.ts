/* eslint-disable @typescript-eslint/no-explicit-any */

/**
 * Provides logging services for a script.
 * 
 * Access the logger through the global `logger` object.
 * 
 * NOTE: In a local test environment, the logger writes to the normal JavaScript.
 * console. In the real Brigade runtime environment, a levelled logger is
 * used which responds appropriately to the configured level.
 */
class Logger {

  /**
   * Logs a message at Debug level.
   * @param message The message to log
   * @param meta Values to replace any substitution strings in `message`
   */
  public debug(message: string, ...meta: any[]): Logger {
    console.debug(message, ...meta)
    return this
  }

  /**
   * Logs a message at Information level.
   * @param message The message to log
   * @param meta Values to replace any substitution strings in `message`
   */
  public info(message: string, ...meta: any[]): Logger {
    console.info(message, ...meta)
    return this
  }

  /**
   * Logs a message at Warning level.
   * @param message The message to log
   * @param meta Values to replace any substitution strings in `message`
   */
  public warn(message: string, ...meta: any[]): Logger {
    console.warn(message, ...meta)
    return this
  }

  /**
   * Logs a message at Error level.
   * @param message The message to log
   * @param meta Values to replace any substitution strings in `message`
   */
  public error(message: string, ...meta: any[]): Logger {
    console.error(message, ...meta)
    return this
  }

}

/** Provides logging for a script. */
export const logger = new Logger()
