"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const cache_1 = require("./cache");
const watch_1 = require("./watch");
exports.ADD = 'add';
exports.UPDATE = 'update';
exports.DELETE = 'delete';
function makeInformer(kubeconfig, path, listPromiseFn) {
    const watch = new watch_1.Watch(kubeconfig);
    return new cache_1.ListWatch(path, watch, listPromiseFn, false);
}
exports.makeInformer = makeInformer;
//# sourceMappingURL=informer.js.map