"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const querystring = require("querystring");
const util_1 = require("util");
const web_socket_handler_1 = require("./web-socket-handler");
class PortForward {
    // handler is a parameter really only for injecting for testing.
    constructor(config, disconnectOnErr, handler) {
        if (!handler) {
            this.handler = new web_socket_handler_1.WebSocketHandler(config);
        }
        else {
            this.handler = handler;
        }
        this.disconnectOnErr = util_1.isUndefined(disconnectOnErr) ? true : disconnectOnErr;
    }
    // TODO: support multiple ports for real...
    portForward(namespace, podName, targetPorts, output, err, input, retryCount = 0) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            if (targetPorts.length === 0) {
                throw new Error('You must provide at least one port to forward to.');
            }
            if (targetPorts.length > 1) {
                throw new Error('Only one port is currently supported for port-forward');
            }
            const query = {
                ports: targetPorts[0],
            };
            const queryStr = querystring.stringify(query);
            const needsToReadPortNumber = [];
            targetPorts.forEach((value, index) => {
                needsToReadPortNumber[index * 2] = true;
                needsToReadPortNumber[index * 2 + 1] = true;
            });
            const path = `/api/v1/namespaces/${namespace}/pods/${podName}/portforward?${queryStr}`;
            const createWebSocket = () => {
                return this.handler.connect(path, null, (streamNum, buff) => {
                    if (streamNum >= targetPorts.length * 2) {
                        return !this.disconnectOnErr;
                    }
                    // First two bytes of each stream are the port number
                    if (needsToReadPortNumber[streamNum]) {
                        buff = buff.slice(2);
                        needsToReadPortNumber[streamNum] = false;
                    }
                    if (streamNum % 2 === 1) {
                        if (err) {
                            err.write(buff);
                        }
                    }
                    else {
                        output.write(buff);
                    }
                    return true;
                });
            };
            if (retryCount < 1) {
                const ws = yield createWebSocket();
                web_socket_handler_1.WebSocketHandler.handleStandardInput(ws, input, 0);
                return ws;
            }
            return web_socket_handler_1.WebSocketHandler.restartableHandleStandardInput(createWebSocket, input, 0, retryCount);
        });
    }
}
exports.PortForward = PortForward;
//# sourceMappingURL=portforward.js.map