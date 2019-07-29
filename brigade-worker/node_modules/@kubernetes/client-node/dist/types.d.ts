import { V1ListMeta, V1ObjectMeta } from './api';
export interface KubernetesObject {
    apiVersion?: string;
    kind?: string;
    metadata?: V1ObjectMeta;
}
export interface KubernetesListObject<T extends KubernetesObject> {
    apiVersion?: string;
    kind?: string;
    metadata?: V1ListMeta;
    items: T[];
}
