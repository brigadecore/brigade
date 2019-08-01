import { Authenticator } from './auth';
import { User } from './config_types';
export declare class ExecAuth implements Authenticator {
    private readonly tokenCache;
    private execFn;
    isAuthProvider(user: User): any;
    getToken(user: User): string | null;
}
