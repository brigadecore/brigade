import { KubeConfig } from './config';
export declare class ProtoClient {
    readonly 'config': KubeConfig;
    get(msgType: any, requestPath: string): Promise<any>;
}
