import { Runnable } from "./runnables"

class Group {
  protected runnables: Runnable[]

  public constructor(...runnables: Runnable[]) {
    this.runnables = runnables || []
  }

  public add(...runnables: Runnable[]): void {
    for (const runnable of runnables) {
      this.runnables.push(runnable)
    }
  }

  public length(): number {
    return this.runnables.length
  }
}

export class SerialGroup extends Group implements Runnable {

  public async run(): Promise<void> {
    for (const runnable of this.runnables) {
      await runnable.run()
    }
  }

}

export class ConcurrentGroup extends Group implements Runnable {

  public async run(): Promise<void> {
    const promises: Promise<void>[] = []
    for (const runnable of this.runnables) {
      promises.push(runnable.run())
    }
    try {
      await Promise.all(promises)
      return Promise.resolve()
    } catch(e) {
      return Promise.reject(e)
    }
  }

}
