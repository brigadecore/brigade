import { User } from './config_types';
export interface Authenticator {
    isAuthProvider(user: User): boolean;
    getToken(user: User): string | null;
}
