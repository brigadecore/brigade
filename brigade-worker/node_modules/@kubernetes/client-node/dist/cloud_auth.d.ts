import { Authenticator } from './auth';
import { User } from './config_types';
export declare class CloudAuth implements Authenticator {
    isAuthProvider(user: User): boolean;
    getToken(user: User): string | null;
    private isExpired;
    private updateAccessToken;
}
