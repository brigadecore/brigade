"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
class OpenIDConnectAuth {
    isAuthProvider(user) {
        if (!user.authProvider) {
            return false;
        }
        return user.authProvider.name === 'oidc';
    }
    getToken(user) {
        if (!user.authProvider.config || !user.authProvider.config['id-token']) {
            return null;
        }
        // TODO: Handle expiration and refresh here...
        // TODO: Extract the 'Bearer ' to config.ts?
        return `Bearer ${user.authProvider.config['id-token']}`;
    }
}
exports.OpenIDConnectAuth = OpenIDConnectAuth;
//# sourceMappingURL=oidc_auth.js.map