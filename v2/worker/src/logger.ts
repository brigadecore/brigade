import * as winston from "winston"
import { TransformableInfo } from "logform"

export const logger = winston.createLogger({
  levels: {
    error: 0,
    warn: 1,
    info: 2,
    debug: 3
  },
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.printf((info: TransformableInfo) => {
      if (info.job) {
        return `${info.timestamp} [job: ${info.job}] ${info.level.toUpperCase()}: ${info.message}`
      }
      return `${info.timestamp} ${info.level.toUpperCase()}: ${info.message}`
    })
  ),
  transports: [ new winston.transports.Console() ]
})
