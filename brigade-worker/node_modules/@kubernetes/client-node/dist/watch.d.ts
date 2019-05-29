import request = require('request');
import { KubeConfig } from './config';
export interface WatchUpdate {
    type: string;
    object: object;
}
export interface RequestInterface {
    webRequest(opts: request.Options, callback: (err: any, response: any, body: any) => void): any;
}
export declare class DefaultRequest implements RequestInterface {
    webRequest(opts: request.Options, callback: (err: any, response: any, body: any) => void): any;
}
export declare class Watch {
    config: KubeConfig;
    private readonly requestImpl;
    constructor(config: KubeConfig, requestImpl?: RequestInterface);
    watch(path: string, queryParams: any, callback: (phase: string, obj: any) => void, done: (err: any) => void): any;
}
