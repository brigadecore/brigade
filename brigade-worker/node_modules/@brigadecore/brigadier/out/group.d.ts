/**
 * group provides features for grouping jobs and managing them collectively.
 */
/** */
import * as jobImpl from "./job";
/**
 * Group describes a collection of associated jobs.
 */
export declare class Group {
    /**
     * runAll is a convenience for running jobs in parallel.
     *
     * This runs a series of jobs in parallel. It is equivalent to
     * `(new Group(jobs)).runAll()`
     */
    static runAll(jobs: jobImpl.Job[]): Promise<jobImpl.Result[]>;
    /**
     * runEach is a convenience of running jobs in a sequence.
     *
     * This runs a series of jobs in order, blocking on each until it completes.
     * It is equivalent to `(new Group(jobs)).runEach()`
     */
    static runEach(jobs: jobImpl.Job[]): Promise<jobImpl.Result[]>;
    protected jobs: jobImpl.Job[];
    constructor(jobs?: jobImpl.Job[]);
    /**
     * add adds one or more jobs to the group.
     */
    add(...j: jobImpl.Job[]): void;
    /**
     * length returns the number of items in the group
     */
    length(): number;
    /**
     * runEach runs each job in order and waits for every one to finish.
     *
     */
    runEach(): Promise<jobImpl.Result[]>;
    /**
     * runAll runs all jobs in parallel and waits for them all to finish.
     */
    runAll(): Promise<jobImpl.Result[]>;
}
