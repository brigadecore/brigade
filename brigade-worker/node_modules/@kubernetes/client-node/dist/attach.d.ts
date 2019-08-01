/// <reference types="node" />
/// <reference types="ws" />
import WebSocket = require('isomorphic-ws');
import stream = require('stream');
import { KubeConfig } from './config';
import { WebSocketInterface } from './web-socket-handler';
export declare class Attach {
    'handler': WebSocketInterface;
    private terminalSizeQueue?;
    constructor(config: KubeConfig, websocketInterface?: WebSocketInterface);
    attach(namespace: string, podName: string, containerName: string, stdout: stream.Writable | any, stderr: stream.Writable | any, stdin: stream.Readable | any, tty: boolean): Promise<WebSocket>;
}
