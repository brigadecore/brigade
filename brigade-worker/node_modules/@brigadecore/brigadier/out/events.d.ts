/**
 * The events package provides event handling support.
 *
 * Brigade scripts are event-driven. Each brigade JS file declares one or more events
 * that it can handle. When the Brigade controller emits a matching event, the
 * appropriate handler is kicked off.
 */
/// <reference types="node" />
import { EventEmitter } from "events";
import { LogLevel } from "./logger";
/**
 * BrigadeEvent describes an event.
 *
 * Brigade is an event-based system. The BrigadeEvent object describes such an
 * event.
 *
 * Every event has a `type` and a `provider`, where the type indicates what
 * sort of event it is (e.g. `push`) and the provider indicates what system
 * provided the event (`github`, `acr`).
 *
 * Most events also have a commit ID, which is associated with the underlying
 * VCS, and a `payload`, which contains the message received from the provider.
 *
 * For example, when a GitHub Push event happens, the BrigadeEvent will have:
 *
 * - type set to `push`
 * - provider set to `github`
 * - commit set to the Git commit ID (e.g. `c0ff3312345...`)
 * - payload set to a string that contains the JSON document received from
 *   GitHub.
 * - buildID set to the build ID.
 *
 * Note that the payload is considered "opaque": It is up to the script to parse
 * it.
 */
export declare class BrigadeEvent {
    /**
     * buildID is the unique ID for this build.
     */
    buildID?: string;
    /**
     * workerID is the ID of the worker responsible for handling this event.
     */
    workerID?: string;
    /**
     * type is the event type ("push", "pull_request")
     */
    type: string;
    /**
     * provider is the thing that triggered the event ("github", "vsts")
     */
    provider?: string;
    /**
     * revision describes a vcs revision.
     */
    revision?: Revision;
    /**
     * payload is the event body.
     * This is the original source from upstream. If upstream returned a string,
     * it is _not_ parsed. For example,
     * if the upstream provider sends a JSON document, this will contain the
     * JSON as a string that must be decoded with something like `JSON.parse()`
     */
    payload?: any;
    /**
     * logLevel is the level at which the Brigade worker will print logs to console.
     * Permitted values are the names of the logLevel enum.
     */
    logLevel: LogLevel;
    cause?: Cause;
}
/**
 * Revision describes a vcs revision.
 */
export declare class Revision {
    /**
     * commit is the upstream VCS commit ID (revision).
     *
     * If it is not provided, it may be interpreted as `master`, or the head
     * of the main branch.
     *
     * The default value is not guaranteed to be `master` in future versions.
     */
    commit?: string;
    ref?: string;
}
/**
 * A Cause is a wrapper around an event. It is used to indicate that this event
 * caused a condition to occur.
 *
 * Frequently this is used to capture a case where an event triggered an error.
 */
export declare class Cause {
    /**
     * The event that was the cause.
     */
    event?: BrigadeEvent;
    /**
     * The reason this event has caused a condition. (Typically, an error object)
     */
    reason?: any;
    /**
     * The mechanism that triggered this event.
     *
     * For example, an exception cather may report "unahndled exception" here.
     */
    trigger?: string;
}
/**
 * Repository describes a source code repository (VCS)
 */
export interface Repository {
    /**
     * name of the repository. For GitHub, this is org/project
     */
    name: string;
    /**
     * cloneURL is the URL at which the repository can be cloned.
     * Traditionally this is https, but with sshKey specified, this can be git+ssh or ssh.
     */
    cloneURL: string;
    /**
     * sshKey the SSH key to use for ssh:// or git+ssh:// protocols
     */
    sshKey?: string;
    /**
     * token is the OAuth2 token for Git interactions over HTTPS
     */
    token?: string;
    /**
     * initGitSubmodules is a flag that controls if the cloned repository should also have it's submodules initialized (if any).
     */
    initGitSubmodules: boolean;
}
/**
 * KubernetesConfig describes Kubernetes configuration for a Project
 */
export interface KubernetesConfig {
    /**
     * namespace is the Kubernetes namespace in which this project should operate.
     */
    namespace: string;
    /**
     * vcsSidecare is the image name for the sidecar container that resolves VCS operations.
     */
    vcsSidecar: string;
    /**
     * vcsSidecarResourcesLimitsCPU is the cpu limits name for the sidecar container.
     */
    vcsSidecarResourcesLimitsCPU: string;
    /**
     * vcsSidecarResourcesLimitsMemory is the memory limits name for the sidecar container.
     */
    vcsSidecarResourcesLimitsMemory: string;
    /**
     * vcsSidecarResourcesRequestsCPU is the cpu requests name for the sidecar container.
     */
    vcsSidecarResourcesRequestsCPU: string;
    /**
     * vcsSidecarResourcesRequestsMemory is the memory requests name for the sidecar container.
     */
    vcsSidecarResourcesRequestsMemory: string;
    /**
     *  buildStorageSize is the size of the build shared storage space used by the jobs
     */
    buildStorageSize: string;
    /**
     * cacheStorageClass is the storage class for job caches.
     */
    cacheStorageClass: string;
    /**
     * buildStorageClass is the storage class for build storage.
     */
    buildStorageClass: string;
}
/**
 * Project represents a Brigade project.
 */
export declare class Project {
    /**
     * id is the unique ID of the project
     */
    id?: string;
    /**
     * name is the project name.
     */
    name?: string;
    /**
     * repo describes the VCS where source for this project can be obtained.
     */
    repo?: Repository;
    /**
     * kubernetes contains the kubernetes configuration for this project.
     */
    kubernetes?: KubernetesConfig;
    secrets: {
        [key: string]: string;
    };
    /**
     * allowPrivilegedJobs enables privileged mode.
     */
    allowPrivilegedJobs: boolean;
    allowHostMounts: boolean;
}
/**
 * EventHandler is an event handler function.
 *
 * An event handler will always receive an event and a project.
 *
 * The EventRegistry.on() method should be supplied a function conformant to EventHandler.
 */
export declare type EventHandler = (e: BrigadeEvent, proj?: Project) => void;
/**
 * EventRegistry manages the registration and execution of events.
 */
export declare class EventRegistry extends EventEmitter {
    /**
     * Create a new event registry.
     */
    constructor();
    has(name: string): boolean;
    /**
     * fire triggers an event.
     * This uses BrigadeEvent.name to fire an event.
     */
    fire(e: BrigadeEvent, proj: Project): void;
    /**
     * Handle a named event using the given EventHandler.
     *
     * While we cannot revise the type that the Node events library takes for callbacks, Brigade will always
     * supply two arguments to the callback, as described in the EventHandler type.
     *
     * @param eventName The name of the event
     * @param cb A callback of the format described in type EventHandler
     */
    on(eventName: string | symbol, cb: ((...args: any[]) => void)): this;
}
