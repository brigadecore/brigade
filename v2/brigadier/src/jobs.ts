import { Event } from "./events"

const defaultTimeout: number = 1000 * 60 * 15

export class Job {
  public name: string
  public primaryContainer: Container
  public sidecarContainers = new Map<string, Container>()
  public timeout: number = defaultTimeout
  public host: JobHost = new JobHost()
  protected event: Event

  constructor(
    name: string,
    image: string,
    event: Event
  ) {
    this.name = name
    this.primaryContainer = new Container(image)
    this.event = event
  }

  public run(): Promise<void> {
    return Promise.resolve()
  }

  public logs(): Promise<string> {
    return Promise.resolve("skipped logs")
  }
}

export class Container {
  public image: string
  public imagePullPolicy = "IfNotPresent"
  public command: string[] = []
  public arguments: string[] = []
  public environment = new Map<string, string>()
  public useWorkspace = false
  public workspaceMountPath = "/var/workspace"
  public useSource = false
  public sourceMountPath = "/var/vcs"
  public privileged = false
  public useHostDockerSocket = false

  constructor(image: string) {
    this.image = image
  }
}

export class JobHost {
  public os?: string
  public nodeSelector: Map<string, string> = new Map<string, string>()
}
