/**
 * The events package provides event handling support.
 *
 * Acid scripts are event-driven. Each acid JS file declares one or more events
 * that it can handle. When the Acid controller emits a matching event, the
 * appropriate handler is kicked off.
 */

/**
 * AcidEvent describes an event.
 *
 * Acid is an event-based system. The AcidEvent object describes such an
 * event.
 *
 * Every event has a `type` and a `provider`, where the type indicates what
 * sort of event it is (e.g. `push`) and the provider indicates what system
 * provided the event (`github`, `acr`).
 *
 * Most events also have a commit ID, which is associated with the underlying
 * VCS, and a `payload`, which contains the message received from the provider.
 *
 * For example, when a GitHub Push event happens, the AcidEvent will have:
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
export class AcidEvent {
    /**
     * buildID is the unique ID for this build.
     */
    buildID: string;
    /**
     * type is the event type ("push", "pull_request")
     */
    type: string;
    /**
     * provider is the thing that triggered the event ("github", "vsts")
     */
    provider: string;
    /**
     * commit is the upstream VCS commit ID (revision).
     *
     * If it is not provided, it may be interpreted as `master`, or the head
     * of the main branch.
     *
     * The default value is not guaranteed to be `master` in future versions.
     */
    commit?: string;
    /**
     * payload is the event body.
     * This is the original source from upstream. If upstream returned a string,
     * it is _not_ parsed. For example,
     * if the upstream provider sends a JSON document, this will contain the
     * JSON as a string that must be decoded with something like `JSON.parse()`
     */
    payload?: any;
    cause?: Cause;
}

/**
 * A Cause is a wrapper around an event. It is used to indicate that this event
 * caused a condition to occur.
 *
 * Frequently this is used to capture a case where an event triggered an error.
 */
export class Cause {
  /**
   * The event that was the cause.
   */
  event: AcidEvent
  /**
   * The reason this event has caused a condition. (Typically, an error object)
   */
  reason?: any
  /**
   * The mechanism that triggered this event.
   *
   * For example, an exception cather may report "unahndled exception" here.
   */
  trigger?: string
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
}

/**
 * Project represents an Acid project.
 */
export class Project {
    /**
     * id is the unique ID of the project
     */
    id: string;
    /**
     * name is the project name.
     */
    name: string;
    /**
     * repo describes the VCS where source for this project can be obtained.
     */
    repo: Repository;
    /**
     * kubernetes contains the kubernetes configuration for this project.
     */
    kubernetes: KubernetesConfig;
    /*
     * secrets is a map of secret names to secret values.
     */
    secrets: {[key:string]: string};
}

/**
 * EventHandler is an event handler function.
 *
 * An event handler will always receive an event and a project.
 */
type EventHandler =  (e: AcidEvent, proj?: Project) => void

/**
 * EventRegistry manages the registration and execution of events.
 */
export class EventRegistry {
  events: {[name: string]: EventHandler}

  /**
   * Create a new event registry.
   */
  constructor() {
      this.events = {"ping": (e: AcidEvent, p: Project) => { console.log("ping") }}
  }

  /**
   * Register an event handler.
   */
  public on(name: string, fn: EventHandler): void {
      this.events[name] = fn
  }

  /**
   * has returns true if the given name has a registered handler.
   */
  public has(name: string): boolean {
    return this.events[name] !== undefined
  }

  /**
   * fire triggers an event.
   * This uses AcidEvent.name to fire an event.
   */
  public fire(e: AcidEvent, proj: Project) {
    if (!this.events[e.type]) {
        return
    }
    let fn = this.events[e.type]
    fn(e, proj)
  }
}
