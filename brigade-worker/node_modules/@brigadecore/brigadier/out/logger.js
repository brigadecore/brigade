"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const PrettyError = require("pretty-error");
const path = require("path");
const rootPath = path.join(__dirname, "..");
const pe = new PrettyError()
    .skipNodeFiles()
    .skipPackage("ts-node")
    .skipPackage("bluebird")
    .alias(rootPath, ".");
pe.start();
var LogLevel;
(function (LogLevel) {
    LogLevel[LogLevel["ALL"] = 0] = "ALL";
    LogLevel[LogLevel["LOG"] = 1] = "LOG";
    LogLevel[LogLevel["INFO"] = 2] = "INFO";
    LogLevel[LogLevel["WARN"] = 3] = "WARN";
    LogLevel[LogLevel["ERROR"] = 4] = "ERROR";
    LogLevel[LogLevel["NONE"] = 5] = "NONE";
})(LogLevel = exports.LogLevel || (exports.LogLevel = {}));
class ContextLogger {
    constructor(ctx = [], logLevel = LogLevel.LOG) {
        if (typeof ctx === "string") {
            ctx = [ctx];
        }
        this.context = `[${new Array("brigade", ...ctx).join(":")}]`;
        this.logLevel = logLevel;
    }
    error(message, ...optionalParams) {
        if (LogLevel.ERROR >= this.logLevel) {
            console.error(this.context, message, ...optionalParams);
        }
    }
    warn(message, ...optionalParams) {
        if (LogLevel.WARN >= this.logLevel) {
            console.warn(this.context, message, ...optionalParams);
        }
    }
    info(message, ...optionalParams) {
        if (LogLevel.INFO >= this.logLevel) {
            console.info(this.context, message, ...optionalParams);
        }
    }
    log(message, ...optionalParams) {
        if (LogLevel.LOG >= this.logLevel) {
            console.log(this.context, message, ...optionalParams);
        }
    }
}
exports.ContextLogger = ContextLogger;
//# sourceMappingURL=logger.js.map