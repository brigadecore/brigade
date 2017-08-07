export class AcidEvent {
    // type is the event type ("push", "pull_request")
    type: string;
    // provider is the thing that triggered the event ("github", "vsts")
    provider: string;
    // commit is the upstream VCS commit ID (revision).
    commit?: string;
    // payload is the event body.
    // This is the original source from upstream.
    // It is a decoded JSON document.
    payload?: any;
}

// Repository describes a source code repository (VCS)
export interface Repository {
    // name of the repository. For GitHub, this is org/project
    name: string;
    // cloneURL is the URL at which the repository can be cloned.
    // Traditionally this is https, but with sshKey specified, this can be git+ssh or ssh.
    cloneURL: string;
    // sshKey the SSH key to use for ssh:// or git+ssh:// protocols
    sshKey?: string;
}

// KubernetesConfig describes Kubernetes configuration for a Project
export interface KubernetesConfig {
    // namespace is the Kubernetes namespace in which this project should operate.
    namespace: string;
    // vcsSidecare is the image name for the sidecar container that resolves VCS operations.
    vcsSidecar: string;
}

// Project represents an Acid project.
export class Project {
    // id is the unique ID of the project
    id: string;
    // name is the project name.
    name: string;
    // repo describes the VCS where source for this project can be obtained.
    repo: Repository;
    // kubernetes contains the kubernetes configuration for this project.
    kubernetes: KubernetesConfig;
    // secrets is a map of secret names to secret values.
    secrets: {[key:string]: string};
}

// EventHandler is an event handler function.
type EventHandler =  (e: AcidEvent, proj?: Project) => void

// EventRegistry manages the registration and execution of events.
export class EventRegistry {
  events: {[name: string]: EventHandler}

  // Create a new event registry.
  constructor() {
      this.events = {"ping": (e: AcidEvent, p: Project) => { console.log("ping") }}
  }

  // Register an event handler.
  public on(name: string, fn: EventHandler): void {
      this.events[name] = fn
  }

  public has(name: string): boolean {
    return this.events[name] !== undefined
  }

  // Trigger an event.
  // This uses AcidEvent.name to fire an event.
  public fire(e: AcidEvent, proj: Project) {
    if (!this.events[e.type]) {
        return
    }
    let fn = this.events[e.type]
    fn(e, proj)
  }
}
