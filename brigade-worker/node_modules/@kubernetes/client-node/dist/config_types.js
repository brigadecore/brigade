"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const fs = require("fs");
const u = require("underscore");
function newClusters(a) {
    return u.map(a, clusterIterator());
}
exports.newClusters = newClusters;
function clusterIterator() {
    return (elt, i, list) => {
        if (!elt.name) {
            throw new Error(`clusters[${i}].name is missing`);
        }
        if (!elt.cluster) {
            throw new Error(`clusters[${i}].cluster is missing`);
        }
        if (!elt.cluster.server) {
            throw new Error(`clusters[${i}].cluster.server is missing`);
        }
        return {
            caData: elt.cluster['certificate-authority-data'],
            caFile: elt.cluster['certificate-authority'],
            name: elt.name,
            server: elt.cluster.server,
            skipTLSVerify: elt.cluster['insecure-skip-tls-verify'] === true,
        };
    };
}
function newUsers(a) {
    return u.map(a, userIterator());
}
exports.newUsers = newUsers;
function userIterator() {
    return (elt, i, list) => {
        if (!elt.name) {
            throw new Error(`users[${i}].name is missing`);
        }
        return {
            authProvider: elt.user ? elt.user['auth-provider'] : null,
            certData: elt.user ? elt.user['client-certificate-data'] : null,
            certFile: elt.user ? elt.user['client-certificate'] : null,
            exec: elt.user ? elt.user.exec : null,
            keyData: elt.user ? elt.user['client-key-data'] : null,
            keyFile: elt.user ? elt.user['client-key'] : null,
            name: elt.name,
            token: findToken(elt.user),
            password: elt.user ? elt.user.password : null,
            username: elt.user ? elt.user.username : null,
        };
    };
}
function findToken(user) {
    if (user) {
        if (user.token) {
            return user.token;
        }
        if (user['token-file']) {
            return fs.readFileSync(user['token-file']).toString();
        }
    }
}
function newContexts(a) {
    return u.map(a, contextIterator());
}
exports.newContexts = newContexts;
function contextIterator() {
    return (elt, i, list) => {
        if (!elt.name) {
            throw new Error(`contexts[${i}].name is missing`);
        }
        if (!elt.context) {
            throw new Error(`contexts[${i}].context is missing`);
        }
        if (!elt.context.cluster) {
            throw new Error(`contexts[${i}].context.cluster is missing`);
        }
        return {
            cluster: elt.context.cluster,
            name: elt.name,
            user: elt.context.user || undefined,
            namespace: elt.context.namespace || undefined,
        };
    };
}
//# sourceMappingURL=config_types.js.map