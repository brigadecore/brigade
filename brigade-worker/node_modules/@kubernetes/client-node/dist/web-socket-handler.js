"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const WebSocket = require("isomorphic-ws");
const protocols = ['v4.channel.k8s.io', 'v3.channel.k8s.io', 'v2.channel.k8s.io', 'channel.k8s.io'];
class WebSocketHandler {
    // factory is really just for test injection
    constructor(config, socketFactory) {
        this.config = config;
        this.socketFactory = socketFactory;
    }
    static handleStandardStreams(streamNum, buff, stdout, stderr) {
        if (buff.length < 1) {
            return null;
        }
        if (stdout && streamNum === WebSocketHandler.StdoutStream) {
            stdout.write(buff);
        }
        else if (stderr && streamNum === WebSocketHandler.StderrStream) {
            stderr.write(buff);
        }
        else if (streamNum === WebSocketHandler.StatusStream) {
            // stream closing.
            if (stdout && stdout !== process.stdout) {
                stdout.end();
            }
            if (stderr && stderr !== process.stderr) {
                stderr.end();
            }
            return JSON.parse(buff.toString('utf8'));
        }
        else {
            throw new Error('Unknown stream: ' + streamNum);
        }
        return null;
    }
    static handleStandardInput(ws, stdin, streamNum = 0) {
        stdin.on('data', (data) => {
            const buff = Buffer.alloc(data.length + 1);
            buff.writeInt8(streamNum, 0);
            if (data instanceof Buffer) {
                data.copy(buff, 1);
            }
            else {
                buff.write(data, 1);
            }
            ws.send(buff);
        });
        stdin.on('end', () => {
            ws.close();
        });
        // Keep the stream open
        return true;
    }
    static restartableHandleStandardInput(createWS, stdin, streamNum = 0, retryCount = 3) {
        if (retryCount < 0) {
            throw new Error("retryCount can't be lower than 0.");
        }
        let queue = Promise.resolve();
        let ws;
        function processData(data) {
            return tslib_1.__awaiter(this, void 0, void 0, function* () {
                const buff = Buffer.alloc(data.length + 1);
                buff.writeInt8(streamNum, 0);
                if (data instanceof Buffer) {
                    data.copy(buff, 1);
                }
                else {
                    buff.write(data, 1);
                }
                let i = 0;
                for (; i < retryCount; ++i) {
                    if (ws !== null && ws.readyState === WebSocket.OPEN) {
                        ws.send(buff);
                        break;
                    }
                    else {
                        ws = yield createWS();
                    }
                }
                if (i >= retryCount) {
                    throw new Error("can't send data to ws");
                }
            });
        }
        stdin.on('data', (data) => {
            queue = queue.then(() => processData(data));
        });
        stdin.on('end', () => {
            if (ws) {
                ws.close();
            }
        });
        return () => ws;
    }
    /**
     * Connect to a web socket endpoint.
     * @param path The HTTP Path to connect to on the server.
     * @param textHandler Callback for text over the web socket.
     *      Returns true if the connection should be kept alive, false to disconnect.
     * @param binaryHandler Callback for binary data over the web socket.
     *      Returns true if the connection should be kept alive, false to disconnect.
     */
    connect(path, textHandler, binaryHandler) {
        const cluster = this.config.getCurrentCluster();
        if (!cluster) {
            throw new Error('No cluster is defined.');
        }
        const server = cluster.server;
        const ssl = server.startsWith('https://');
        const target = ssl ? server.substr(8) : server.substr(7);
        const proto = ssl ? 'wss' : 'ws';
        const uri = `${proto}://${target}${path}`;
        const opts = {};
        this.config.applytoHTTPSOptions(opts);
        return new Promise((resolve, reject) => {
            const client = this.socketFactory
                ? this.socketFactory(uri, opts)
                : new WebSocket(uri, protocols, opts);
            let resolved = false;
            client.onopen = () => {
                resolved = true;
                resolve(client);
            };
            client.onerror = (err) => {
                if (!resolved) {
                    reject(err);
                }
            };
            client.onmessage = ({ data }) => {
                // TODO: support ArrayBuffer and Buffer[] data types?
                if (typeof data === 'string') {
                    if (textHandler && !textHandler(data)) {
                        client.close();
                    }
                }
                else if (data instanceof Buffer) {
                    const streamNum = data.readInt8(0);
                    if (binaryHandler && !binaryHandler(streamNum, data.slice(1))) {
                        client.close();
                    }
                }
            };
        });
    }
}
WebSocketHandler.StdinStream = 0;
WebSocketHandler.StdoutStream = 1;
WebSocketHandler.StderrStream = 2;
WebSocketHandler.StatusStream = 3;
WebSocketHandler.ResizeStream = 4;
exports.WebSocketHandler = WebSocketHandler;
//# sourceMappingURL=web-socket-handler.js.map