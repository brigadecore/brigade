/// <reference types="node" />
import https = require('https');
import request = require('request');
import * as api from './api';
import { Cluster, Context, User } from './config_types';
export declare class KubeConfig {
    private static authenticators;
    /**
     * The list of all known clusters
     */
    'clusters': Cluster[];
    /**
     * The list of all known users
     */
    'users': User[];
    /**
     * The list of all known contexts
     */
    'contexts': Context[];
    /**
     * The name of the current context
     */
    'currentContext': string;
    getContexts(): Context[];
    getClusters(): Cluster[];
    getUsers(): User[];
    getCurrentContext(): string;
    setCurrentContext(context: string): void;
    getContextObject(name: string): Context | null;
    getCurrentCluster(): Cluster | null;
    getCluster(name: string): Cluster | null;
    getCurrentUser(): User | null;
    getUser(name: string): User | null;
    loadFromFile(file: string): void;
    applytoHTTPSOptions(opts: https.RequestOptions): void;
    applyToRequest(opts: request.Options): void;
    loadFromString(config: string): void;
    loadFromOptions(options: any): void;
    loadFromClusterAndUser(cluster: Cluster, user: User): void;
    loadFromCluster(pathPrefix?: string): void;
    mergeConfig(config: KubeConfig): void;
    addCluster(cluster: Cluster): void;
    addUser(user: User): void;
    addContext(ctx: Context): void;
    loadFromDefault(): void;
    makeApiClient<T extends ApiType>(apiClientType: ApiConstructor<T>): T;
    makePathsAbsolute(rootDirectory: string): void;
    private getCurrentContextObject;
    private applyHTTPSOptions;
    private applyAuthorizationHeader;
    private applyOptions;
}
export interface ApiType {
    setDefaultAuthentication(config: api.Authentication): any;
}
declare type ApiConstructor<T extends ApiType> = new (server: string) => T;
export declare class Config {
    static SERVICEACCOUNT_ROOT: string;
    static SERVICEACCOUNT_CA_PATH: string;
    static SERVICEACCOUNT_TOKEN_PATH: string;
    static fromFile(filename: string): api.CoreV1Api;
    static fromCluster(): api.CoreV1Api;
    static defaultClient(): api.CoreV1Api;
    static apiFromFile<T extends ApiType>(filename: string, apiClientType: ApiConstructor<T>): T;
    static apiFromCluster<T extends ApiType>(apiClientType: ApiConstructor<T>): T;
    static apiFromDefaultClient<T extends ApiType>(apiClientType: ApiConstructor<T>): T;
}
export declare function makeAbsolutePath(root: string, file: string): string;
export declare function bufferFromFileOrString(file?: string, data?: string): Buffer | null;
export declare function findHomeDir(): string | null;
export interface Named {
    name: string;
}
export declare function findObject<T extends Named>(list: T[], name: string, key: string): T | null;
export {};
