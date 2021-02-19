import { Event } from "./events"
import { Runnable } from "./runnables"

const defaultTimeout: number = 1000 * 60 * 15

export class Job implements Runnable {
  public name: string
  public primaryContainer: Container
  public sidecarContainers: { [key: string]: Container } = {}
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
  public workingDirectory = ""
  public command: string[] = []
  public arguments: string[] = []
  public environment: { [key: string]: string } = {}
  public workspaceMountPath = ""
  public sourceMountPath = ""
  public privileged = false
  public useHostDockerSocket = false

  constructor(image: string) {
    this.image = image
  }
}

export class JobHost {
  public os?: string
  public nodeSelector: { [key: string]: string } = {}
}
