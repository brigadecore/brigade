"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const byline_1 = require("byline");
const request = require("request");
class DefaultRequest {
    webRequest(opts, callback) {
        return request(opts, callback);
    }
}
exports.DefaultRequest = DefaultRequest;
class Watch {
    constructor(config, requestImpl) {
        this.config = config;
        if (requestImpl) {
            this.requestImpl = requestImpl;
        }
        else {
            this.requestImpl = new DefaultRequest();
        }
    }
    watch(path, queryParams, callback, done) {
        const cluster = this.config.getCurrentCluster();
        if (!cluster) {
            throw new Error('No currently active cluster');
        }
        const url = cluster.server + path;
        queryParams.watch = true;
        const headerParams = {};
        const requestOptions = {
            method: 'GET',
            qs: queryParams,
            headers: headerParams,
            uri: url,
            useQuerystring: true,
            json: true,
        };
        this.config.applyToRequest(requestOptions);
        const stream = new byline_1.LineStream();
        stream.on('data', (data) => {
            let obj;
            if (data instanceof Buffer) {
                obj = JSON.parse(data.toString());
            }
            else {
                obj = JSON.parse(data);
            }
            if (typeof obj === 'object' && obj.object) {
                callback(obj.type, obj.object);
            }
            else {
                throw new Error(`unexpected ${typeof obj}: ${JSON.stringify(obj)}`);
            }
        });
        const req = this.requestImpl.webRequest(requestOptions, (error, response, body) => {
            if (error) {
                done(error);
            }
            else {
                done(null);
            }
        });
        req.pipe(stream);
        return req;
    }
}
exports.Watch = Watch;
//# sourceMappingURL=watch.js.map