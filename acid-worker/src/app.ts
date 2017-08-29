/**
 * Module app is the main application runner.
 */
import * as ulid from "ulid"

import * as events from "./events"
import * as process from "process"
import * as k8s from "./k8s"
import * as libacid from './libacid'

interface BuildStorage {
  create(buildID: string, project: events.Project, size?: string): Promise<string>
  destroy(): Promise<boolean>
}

/**
 * ProjectLoader describes a function able to load a Project.
 */
interface ProjectLoader {
  (projectID: string, projectNS: string): Promise<events.Project>
}

/**
 * App is the main application.
 *
 * App assumes that it has full control of the process. It acts as a top-level
 * error handler and will exit the process with errors when uncaught resolutions
 * and errors occur.
 */
export class App {
  /**
   * exitOnError controls whether the app will exit when an uncaught exception or unhandled rejection occurs.
   *
   * exitOnError can be set to false in order to run tests on the error handling.
   * In general, though, it should be left on. In some cases, by the time the
   * process trap is invoked, the runtime is not in a good state to continue.
   */
  public exitOnError: boolean = true
  protected errorsHandled: boolean = false
  protected lastEvent: events.AcidEvent
  protected projectID: string
  protected projectNS: string
  // On project loading error, this value may be passed. In all other cases,
  // it is overwritten by an actual project.
  protected proj: events.Project = new events.Project()

  // true if the "after" event has fired.
  protected afterHasFired: boolean = false

  /**
   * loadProject is a function that loads projects.
   */
  public loadProject: ProjectLoader = k8s.loadProject
  /**
   * buildStorage controls the per-build storage layer.
   */
  public buildStorage: BuildStorage = new k8s.BuildStorage()


  /**
   * Create a new App.
   *
   * An app requires a project ID and project NS.
   */
  constructor(projectID: string, projectNS: string) {
    this.projectID = projectID
    this.projectNS = projectNS
  }
  /**
   * run runs a particular event for this app.
   */
  public run(e: events.AcidEvent): Promise<boolean> {
    this.lastEvent = e

    // Run if an uncaught rejection happens.
    process.on("unhandledRejection", (reason: any, p: Promise<any>) => {
      var msg = reason
      // Kubernetes objects put error messages here:
      if (reason.body && reason.body.message) {
        msg = reason.body.message
      }
      console.log(`FATAL: ${ msg } (rejection)`)
      this.fireError(reason, "unhandledRejection")
      this.exitOnError && process.exit(3)
      this.buildStorage.destroy().then( (destroyed) => {
        if (!destroyed) {
          console.log(`storage not destroyed for ${ this.proj.name }`)
        }
        this.exitOnError && process.exit(3)
      }).catch( (reason) => {
        console.log(`error prevented storage cleanup for ${ this.proj.name }: ${ reason }`)
        this.exitOnError && process.exit(3)
      })
    })

    // Run at the end.
    process.on("beforeExit", code => {
      if (this.afterHasFired) {
        return
      }
      this.afterHasFired = true

      let after: events.AcidEvent = {
        buildID: e.buildID,
        type: "after",
        provider: "acid",
        commit: e.commit,
        cause: {
          event: e,
          trigger: code == 0 ? "success" : "failure"
        } as events.Cause
      }
      libacid.fire(after, this.proj)
      // Teardown storage.
      // Note that there is no guarantee here that the storage will be available
      // inside of the "after" event handler. Because "after" is async, the storage
      // can be torn down before the "after" handlers are executed.
      this.buildStorage.destroy()
    })

    // TODO: fire() should also return a promise, and that promise's result
    // should be bubbled up.
    return new Promise( (resolve, reject) => {
      // This traps unhandled 'throw' calls, and is considered safer than
      // process.on("unhandledException"). In most cases, the unhandledRejection
      // handler will trigger before this does.
      // TODO: Can we remove the try/catch block?
      try {
        // Load the project, then fire the event, then fire the "after" event.
        this.loadProject(this.projectID, this.projectNS).then (p => {
          this.proj = p
          // Setup storage
          return this.buildStorage.create(e.buildID.toLowerCase(), p, "50Mi")
        }).then( () => {
          libacid.fire(e, this.proj)
        }) // We want the missing catch() to trigger the main rejection handler.
      } catch (e) {
        console.log(`FATAL: ${ e } (exception)`)
        this.fireError(e, "uncaughtException")

        this.buildStorage.destroy().then( (destroyed) => {
          if (!destroyed) {
            console.log(`storage not destroyed for ${ this.proj.name }`)
          }
          this.exitOnError && process.exit(3)
          reject(false)
        }).catch( (reason) => {
          console.log(`error prevented storage cleanup for ${ this.proj.name }: ${ reason }`)
          this.exitOnError && process.exit(3)
          reject(false)
        })
      }
      resolve(true)
    })
  }

  /**
   * fireError fires an "error" event when the top-level script catches an error.
   *
   * It is fired no more than once, and is only fired when the error bubbles all
   * the way to the top.
   */
  public fireError(reason?: any, errorType?: string): void {
    if (this.errorsHandled) {
      return
    }
    this.errorsHandled = true

    let errorEvent: events.AcidEvent = {
      buildID: this.lastEvent.buildID,
      type: "error",
      provider: "acid",
      commit: this.lastEvent.commit,
      cause: {
        event: this.lastEvent,
        reason: reason,
        trigger: errorType
      } as events.Cause
    }

    libacid.fire(errorEvent, this.proj)
  }

  /**
   * Generate a random build ID.
   */
  public static generateBuildID(commit: string): string {
    return `acid-worker-${ ulid() }-${ commit.substring(0, 8) }`
  }
}

