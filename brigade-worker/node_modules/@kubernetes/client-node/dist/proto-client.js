"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const http = require("http");
const url = require("url");
class ProtoClient {
    get(msgType, requestPath) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            const server = this.config.getCurrentCluster().server;
            const u = new url.URL(server);
            const options = {
                path: requestPath,
                hostname: u.hostname,
                protocol: u.protocol,
            };
            this.config.applytoHTTPSOptions(options);
            const req = http.request(options);
            const result = new Promise((resolve, reject) => {
                let data = '';
                req.on('data', (chunk) => {
                    data = data + chunk;
                });
                req.on('end', () => {
                    const obj = msgType.deserializeBinary(data);
                    resolve(obj);
                });
                req.on('error', (err) => {
                    reject(err);
                });
            });
            req.end();
            return result;
        });
    }
}
exports.ProtoClient = ProtoClient;
//# sourceMappingURL=proto-client.js.map