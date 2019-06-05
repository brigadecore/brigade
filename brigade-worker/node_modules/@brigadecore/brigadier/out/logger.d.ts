export declare enum LogLevel {
    ALL = 0,
    LOG = 1,
    INFO = 2,
    WARN = 3,
    ERROR = 4,
    NONE = 5
}
export interface Logger {
    logLevel: LogLevel;
    error(message?: any, ...optionalParams: any[]): void;
    warn(message?: any, ...optionalParams: any[]): void;
    info(message?: any, ...optionalParams: any[]): void;
    log(message?: any, ...optionalParams: any[]): void;
}
export declare class ContextLogger implements Logger {
    context: string;
    logLevel: LogLevel;
    constructor(ctx?: string[] | string, logLevel?: LogLevel);
    error(message?: any, ...optionalParams: any[]): void;
    warn(message?: any, ...optionalParams: any[]): void;
    info(message?: any, ...optionalParams: any[]): void;
    log(message?: any, ...optionalParams: any[]): void;
}
