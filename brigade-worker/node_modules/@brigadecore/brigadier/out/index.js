"use strict";
/**
 * brigadier is the client-facing library for Brigade.
 *
 * Objects in this library are available to `brigade.js` scripts.
 */
Object.defineProperty(exports, "__esModule", { value: true });
/** */
const jobImpl = require("./job");
const groupImpl = require("./group");
const eventsImpl = require("./events");
/**
 * events is the main event registry.
 *
 * New event handlers can be registered using `events.on(name: string, (e: BrigadeEvent,
 * p: Project) => {})`.
 * where the `name` is the event name, and the callback is the function to be
 * executed when the event is triggered.
 */
exports.events = new eventsImpl.EventRegistry();
/**
 * fire triggers an event.
 *
 * The fire() function takes a BrigadeEvent (the event to be triggered) and a
 * Project (the owner project). If an event handler is found, it is executed.
 * If no event handler is found, nothing happens.
 */
function fire(e, p) {
    exports.events.fire(e, p);
}
exports.fire = fire;
/**
 * Job describes a particular job.
 *
 * A Job always has a name and an image. The name is used to reference this
 * job in relation to other jobs in the same event. The image corresponds to a
 * container image that will be executed as part of this job.
 *
 * A Job may also have one or more tasks associated with it. Tasks are run
 * (in order) inside of the image. When no tasks are supplied, the image is
 * executed as-is.
 *
 * The version of Job that ships with this package has no runtime. Jobs are
 */
class Job extends jobImpl.Job {
    constructor() {
        super(...arguments);
        this.runResponse = "skipped run";
        this.logsResponse = "skipped logs";
    }
    run() {
        return Promise.resolve(this.runResponse);
    }
    logs() {
        return Promise.resolve(this.logsResponse);
    }
}
exports.Job = Job;
/**
 * Group describes a collection of associated jobs.
 *
 * A group of jobs can be executed in two ways:
 *   - In parallel as runAll()
 *   - In serial as runEach()
 */
class Group extends groupImpl.Group {
}
exports.Group = Group;
/**
 * ErrorReport describes an error in the runtime handling of a Brigade script.
 */
class ErrorReport {
}
exports.ErrorReport = ErrorReport;
//# sourceMappingURL=index.js.map