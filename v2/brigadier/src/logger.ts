class Logger {

  public debug(message: string, ...meta: any[]): Logger {
    console.debug(message, ...meta)
    return this
  }

  public info(message: string, ...meta: any[]): Logger {
    console.info(message, ...meta)
    return this
  }

  public warn(message: string, ...meta: any[]): Logger {
    console.warn(message, ...meta)
    return this
  }

  public error(message: string, ...meta: any[]): Logger {
    console.error(message, ...meta)
    return this
  }

}

export const logger = new Logger()
