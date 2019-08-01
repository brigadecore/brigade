/// <reference types="node" />
/// <reference types="ws" />
import WebSocket = require('isomorphic-ws');
import stream = require('stream');
import { V1Status } from './api';
import { KubeConfig } from './config';
export interface WebSocketInterface {
    connect(path: string, textHandler: ((text: string) => boolean) | null, binaryHandler: ((stream: number, buff: Buffer) => boolean) | null): Promise<WebSocket>;
}
export declare class WebSocketHandler implements WebSocketInterface {
    readonly config: KubeConfig;
    readonly socketFactory?: ((uri: string, opts: WebSocket.ClientOptions) => WebSocket) | undefined;
    static readonly StdinStream = 0;
    static readonly StdoutStream = 1;
    static readonly StderrStream = 2;
    static readonly StatusStream = 3;
    static readonly ResizeStream = 4;
    static handleStandardStreams(streamNum: number, buff: Buffer, stdout: stream.Writable | null, stderr: stream.Writable | null): V1Status | null;
    static handleStandardInput(ws: WebSocket, stdin: stream.Readable | any, streamNum?: number): boolean;
    static restartableHandleStandardInput(createWS: () => Promise<WebSocket>, stdin: stream.Readable | any, streamNum?: number, retryCount?: number): () => WebSocket | null;
    constructor(config: KubeConfig, socketFactory?: ((uri: string, opts: WebSocket.ClientOptions) => WebSocket) | undefined);
    /**
     * Connect to a web socket endpoint.
     * @param path The HTTP Path to connect to on the server.
     * @param textHandler Callback for text over the web socket.
     *      Returns true if the connection should be kept alive, false to disconnect.
     * @param binaryHandler Callback for binary data over the web socket.
     *      Returns true if the connection should be kept alive, false to disconnect.
     */
    connect(path: string, textHandler: ((text: string) => boolean) | null, binaryHandler: ((stream: number, buff: Buffer) => boolean) | null): Promise<WebSocket>;
}
