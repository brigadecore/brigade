"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const querystring = require("querystring");
const terminal_size_queue_1 = require("./terminal-size-queue");
const web_socket_handler_1 = require("./web-socket-handler");
class Exec {
    constructor(config, wsInterface) {
        if (wsInterface) {
            this.handler = wsInterface;
        }
        else {
            this.handler = new web_socket_handler_1.WebSocketHandler(config);
        }
    }
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
    exec(namespace, podName, containerName, command, stdout, stderr, stdin, tty, statusCallback) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            const query = {
                stdout: stdout != null,
                stderr: stderr != null,
                stdin: stdin != null,
                tty,
                command,
                container: containerName,
            };
            const queryStr = querystring.stringify(query);
            const path = `/api/v1/namespaces/${namespace}/pods/${podName}/exec?${queryStr}`;
            const conn = yield this.handler.connect(path, null, (streamNum, buff) => {
                const status = web_socket_handler_1.WebSocketHandler.handleStandardStreams(streamNum, buff, stdout, stderr);
                if (status != null) {
                    if (statusCallback) {
                        statusCallback(status);
                    }
                    return false;
                }
                return true;
            });
            if (stdin != null) {
                web_socket_handler_1.WebSocketHandler.handleStandardInput(conn, stdin, web_socket_handler_1.WebSocketHandler.StdinStream);
            }
            if (terminal_size_queue_1.isResizable(stdout)) {
                this.terminalSizeQueue = new terminal_size_queue_1.TerminalSizeQueue();
                web_socket_handler_1.WebSocketHandler.handleStandardInput(conn, this.terminalSizeQueue, web_socket_handler_1.WebSocketHandler.ResizeStream);
                this.terminalSizeQueue.handleResizes(stdout);
            }
            return conn;
        });
    }
}
exports.Exec = Exec;
//# sourceMappingURL=exec.js.map