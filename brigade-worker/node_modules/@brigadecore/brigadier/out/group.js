"use strict";
/**
 * group provides features for grouping jobs and managing them collectively.
 */
Object.defineProperty(exports, "__esModule", { value: true });
/**
 * Group describes a collection of associated jobs.
 */
class Group {
    constructor(jobs) {
        this.jobs = [];
        this.jobs = jobs || [];
    }
    /**
     * runAll is a convenience for running jobs in parallel.
     *
     * This runs a series of jobs in parallel. It is equivalent to
     * `(new Group(jobs)).runAll()`
     */
    static runAll(jobs) {
        let g = new Group(jobs);
        return g.runAll();
    }
    /**
     * runEach is a convenience of running jobs in a sequence.
     *
     * This runs a series of jobs in order, blocking on each until it completes.
     * It is equivalent to `(new Group(jobs)).runEach()`
     */
    static runEach(jobs) {
        let g = new Group(jobs);
        return g.runEach();
    }
    /**
     * add adds one or more jobs to the group.
     */
    add(...j) {
        for (let jj of j) {
            this.jobs.push(jj);
        }
    }
    /**
     * length returns the number of items in the group
     */
    length() {
        return this.jobs.length;
    }
    /**
     * runEach runs each job in order and waits for every one to finish.
     *
     */
    runEach() {
        // TODO: Rewrite this using async/await, which will make it much cleaner.
        return this.jobs.reduce((promise, job) => {
            return promise.then((results) => {
                return job.run().then(jobResult => {
                    results.push(jobResult);
                    return results;
                });
            });
        }, Promise.resolve([]));
    }
    /**
     * runAll runs all jobs in parallel and waits for them all to finish.
     */
    runAll() {
        let plist = [];
        for (let j of this.jobs) {
            plist.push(j.run());
        }
        return Promise.all(plist);
    }
}
exports.Group = Group;
//# sourceMappingURL=group.js.map