/// <reference types="node" />
/// <reference types="ws" />
import WebSocket = require('isomorphic-ws');
import stream = require('stream');
import { V1Status } from './api';
import { KubeConfig } from './config';
import { WebSocketInterface } from './web-socket-handler';
export declare class Exec {
    'handler': WebSocketInterface;
    private terminalSizeQueue?;
    constructor(config: KubeConfig, wsInterface?: WebSocketInterface);
    /**
     * @param {string}  namespace - The namespace of the pod to exec the command inside.
     * @param {string} podName - The name of the pod to exec the command inside.
     * @param {string} containerName - The name of the container in the pod to exec the command inside.
     * @param {(string|string[])} command - The command or command and arguments to execute.
     * @param {stream.Writable} stdout - The stream to write stdout data from the command.
     * @param {stream.Writable} stderr - The stream to write stderr data from the command.
     * @param {stream.Readable} stdin - The strream to write stdin data into the command.
     * @param {boolean} tty - Should the command execute in a TTY enabled session.
     * @param {(V1Status) => void} statusCallback -
     *       A callback to received the status (e.g. exit code) from the command, optional.
     * @return {string} This is the result
     */
    exec(namespace: string, podName: string, containerName: string, command: string | string[], stdout: stream.Writable | null, stderr: stream.Writable | null, stdin: stream.Readable | null, tty: boolean, statusCallback?: (status: V1Status) => void): Promise<WebSocket>;
}
