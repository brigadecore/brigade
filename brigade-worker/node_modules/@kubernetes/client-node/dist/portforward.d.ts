/// <reference types="node" />
/// <reference types="ws" />
import WebSocket = require('isomorphic-ws');
import stream = require('stream');
import { KubeConfig } from './config';
import { WebSocketInterface } from './web-socket-handler';
export declare class PortForward {
    private readonly handler;
    private readonly disconnectOnErr;
    constructor(config: KubeConfig, disconnectOnErr?: boolean, handler?: WebSocketInterface);
    portForward(namespace: string, podName: string, targetPorts: number[], output: stream.Writable, err: stream.Writable | null, input: stream.Readable, retryCount?: number): Promise<WebSocket | (() => WebSocket | null)>;
}
