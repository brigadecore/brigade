/**
 * Module app is the main application runner.
 */
import * as events from "@azure/brigadier/out/events";
import * as process from "process";
import * as k8s from "./k8s";
import * as brigadier from "./brigadier";
import { Logger, ContextLogger } from "@azure/brigadier/out/logger";

interface BuildStorage {
  create(
    e: events.BrigadeEvent,
    project: events.Project,
    size?: string
  ): Promise<string>;
  destroy(): Promise<boolean>;
}

/**
 * ProjectLoader describes a function able to load a Project.
 */
type ProjectLoader = (
  projectID: string,
  projectNS: string
) => Promise<events.Project>;

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
  public exitOnError: boolean = true;
  protected errorsHandled: boolean = false;
  protected logger: Logger = new ContextLogger("app");
  protected lastEvent: events.BrigadeEvent;
  protected projectID: string;
  protected projectNS: string;
  // On project loading error, this value may be passed. In all other cases,
  // it is overwritten by an actual project.
  protected proj: events.Project = new events.Project();

  // true if the "after" event has fired.
  protected afterHasFired: boolean = false;
  protected storageIsDestroyed: boolean = false;
  /**
   * loadProject is a function that loads projects.
   */
  public loadProject: ProjectLoader = k8s.loadProject;
  /**
   * buildStorage controls the per-build storage layer.
   */
  public buildStorage: BuildStorage = new k8s.BuildStorage();

  protected exitCode: number = 0;

  /**
   * Create a new App.
   *
   * An app requires a project ID and project NS.
   */
  constructor(projectID: string, projectNS: string) {
    this.projectID = projectID;
    this.projectNS = projectNS;
  }

  /**
   * run runs a particular event for this app.
   */
  public run(e: events.BrigadeEvent): Promise<boolean> {
    this.lastEvent = e;
    this.logger.logLevel = e.logLevel;

    // This closure destroys storage for us. It is called by event handlers.
    let destroyStorage = () => {
      // Since we catch a destroy error, the outer wrapper will
      // not get that error. Essentially, we swallow the error to prevent
      // cleanup from exiting > 0.
      return this.buildStorage
        .destroy()
        .then(destroyed => {
          if (!destroyed) {
            this.logger.log(`storage not destroyed for ${e.workerID}`);
          }
        })
        .catch(reason => {
          // Kubernetes objects put error messages here:
          const msg = reason.body ? reason.body.message : reason;
          this.logger.log(
            `failed to destroy storage for ${e.workerID}: ${msg}`
          );
        });
    };

    // We need at least one error trap to avoid losing the error to a new
    // throw from EventEmitter.
    brigadier.events.once("error", () => {
      this.logger.log("error handler is cleaning up");
      if (this.exitOnError) {
        destroyStorage().then(() => {
          // Docs say this should work, but it produces an tsc error.
          // process.exitCode = 1
          // So we work around by storing state and calling process.exit()
          // in the "exit" event handler.
          this.exitCode = 1;
        });
      }
    });

    // We need to ensure that after is called exactly once. So we need an
    // empty after handler.
    brigadier.events.once("after", () => {
      this.afterHasFired = true;

      // Delay long enough to cause beforeExit to be emitted again.
      setImmediate(() => {
        this.logger.info("after: default event handler fired");
      }, 20);
    });

    // Run if an uncaught rejection happens.
    process.on("unhandledRejection", (reason: any, p: Promise<any>) => {
      this.logger.error(reason);
      this.fireError(reason, "unhandledRejection");
    });

    process.on("exit", code => {
      if (this.exitCode != 0) {
        process.exit(this.exitCode);
      }
    });

    // Run at the end.
    process.on("beforeExit", code => {
      // If an error has occurred, skip running the after handler.
      if (this.exitCode != 0) {
        return;
      }

      if (this.afterHasFired) {
        // So at this point, the after event has fired and we can cleanup.
        if (!this.storageIsDestroyed) {
          this.logger.log("beforeExit(2): destroying storage");
          this.storageIsDestroyed = true;
          destroyStorage();
        }
        return;
      }

      let after: events.BrigadeEvent = {
        buildID: e.buildID,
        workerID: e.workerID,
        type: "after",
        provider: "brigade",
        revision: e.revision,
        logLevel: e.logLevel,
        cause: {
          event: e,
          trigger: code == 0 ? "success" : "failure"
        } as events.Cause
      };

      // Only fire an event if the top-level had a match.
      if (brigadier.events.has(e.type)) {
        brigadier.fire(after, this.proj);
      } else {
        this.afterHasFired = true;
        setImmediate(() => {
          this.logger.log("no-after: fired");
        }, 20);
      }
    });

    // Now that we have all the handlers registered, load the project and
    // execute the event.
    return this.loadProject(this.projectID, this.projectNS)
      .then(p => {
        this.proj = p;
        // Setup storage
        return this.buildStorage.create(e, p, p.kubernetes.buildStorageSize);
      })
      .then(() => {
        brigadier.fire(e, this.proj);
        return true;
      }); // We want to trigger the main rejection handler, so we do not catch().
  }

  /**
   * fireError fires an "error" event when the top-level script catches an error.
   *
   * It is fired no more than once, and is only fired when the error bubbles all
   * the way to the top.
   */
  public fireError(reason?: any, errorType?: string): void {
    if (this.errorsHandled) {
      return;
    }
    this.errorsHandled = true;

    let errorEvent: events.BrigadeEvent = {
      buildID: this.lastEvent.buildID,
      workerID: this.lastEvent.workerID,
      type: "error",
      provider: "brigade",
      revision: this.lastEvent.revision,
      logLevel: this.lastEvent.logLevel,
      cause: {
        event: this.lastEvent,
        reason: reason,
        trigger: errorType
      } as events.Cause
    };

    brigadier.fire(errorEvent, this.proj);
  }
}
