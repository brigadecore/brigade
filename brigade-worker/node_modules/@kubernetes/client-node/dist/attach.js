"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const querystring = require("querystring");
const terminal_size_queue_1 = require("./terminal-size-queue");
const web_socket_handler_1 = require("./web-socket-handler");
class Attach {
    constructor(config, websocketInterface) {
        if (websocketInterface) {
            this.handler = websocketInterface;
        }
        else {
            this.handler = new web_socket_handler_1.WebSocketHandler(config);
        }
    }
    attach(namespace, podName, containerName, stdout, stderr, stdin, tty) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            const query = {
                container: containerName,
                stderr: stderr != null,
                stdin: stdin != null,
                stdout: stdout != null,
                tty,
            };
            const queryStr = querystring.stringify(query);
            const path = `/api/v1/namespaces/${namespace}/pods/${podName}/attach?${queryStr}`;
            const conn = yield this.handler.connect(path, null, (streamNum, buff) => {
                web_socket_handler_1.WebSocketHandler.handleStandardStreams(streamNum, buff, stdout, stderr);
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
exports.Attach = Attach;
//# sourceMappingURL=attach.js.map