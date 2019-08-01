"use strict";
/**
 * Package job provides support for jobs.
 *
 * A Job idescribes a particular unit of a build. A Job returns a Result.
 * A JobRunner is an implementation of the runtime logic for a Job.
 */
Object.defineProperty(exports, "__esModule", { value: true });
/**
 * The default shell for the job.
 */
const defaultShell = "/bin/sh";
/**
 * defaultTimeout is the default timeout for a job (15 minutes)
 */
const defaultTimeout = 1000 * 60 * 15;
/**
 * The default image if `Job.image` is not set
 */
const brigadeImage = "debian:jessie-slim";
exports.brigadeCachePath = "/mnt/brigade/cache";
exports.brigadeStoragePath = "/mnt/brigade/share";
exports.dockerSocketMountPath = "/var/run/docker.sock";
exports.dockerSocketMountName = "docker-socket";
/**
 * Cache controls the job's cache.
 *
 * A cache is a small storage space that is shared between different instances
 * if the same job.
 *
 * Cache is just a plain filesystem, and as such comes with no guarantees about
 * consistency, etc. It should be treated as volatile.
 */
class JobCache {
    constructor() {
        /**
         * If enabled=true, a storage cache will be attached.
         */
        this.enabled = false;
        /**
         * size is the amount of storage space assigned to the cache. The default is
         * 5Mi.
         * For sizing information, see https://github.com/kubernetes/community/blob/master/contributors/design-proposals/resources.md
         */
        this.size = "5Mi";
        // EXPERIMENTAL: Allow script authors to change this location.
        // Before Brigade 0.15, this used a getter to prevent scripters from setting
        // this path directly.
        this.path = exports.brigadeCachePath;
    }
}
exports.JobCache = JobCache;
/**
 * JobStorage configures build-wide storage preferences for this job.
 *
 * Changes to this object only impact the job, not the entire build.
 */
class JobStorage {
    constructor() {
        this.enabled = false;
        // EXPERIMENTAL: Allow setting the path.
        // Prior to Brigade 0.15, this was read-only.
        this.path = exports.brigadeStoragePath;
    }
}
exports.JobStorage = JobStorage;
/**
 * JobHost expresses expectations about the host a job will run on.
 */
class JobHost {
    constructor() {
        this.nodeSelector = new Map();
    }
}
exports.JobHost = JobHost;
/**
 * JobDockerMount enables or disables mounting the host's docker socket for a job.
 */
class JobDockerMount {
    constructor() {
        /**
         * enabled configues whether or not the job will mount the host's docker socket.
         */
        this.enabled = false;
    }
}
exports.JobDockerMount = JobDockerMount;
/**
 * JobResourceRequest represents request of the resources
 */
class JobResourceRequest {
}
exports.JobResourceRequest = JobResourceRequest;
/**
 * JobResourceLimit represents limit of the resources
 */
class JobResourceLimit {
}
exports.JobResourceLimit = JobResourceLimit;
/**
 * Job represents a single job, which is composed of several closely related sequential tasks.
 * Jobs must have names. Every job also has an associated image, which references
 * the Docker container to be run.
 * */
class Job {
    /** Create a new Job
     * name is the name of the job.
     * image is the container image to use
     * tasks is a list of commands to run.
     */
    constructor(name, image, tasks, imageForcePull = false) {
        /** shell that will be used by default in this job*/
        this.shell = defaultShell;
        /** image is the container image to be run*/
        this.image = brigadeImage;
        /** imageForcePull defines the container image pull policy: Always if true or IfNotPresent if false */
        this.imageForcePull = false;
        /**
         * imagePullSecrets names secrets that contain the credentials for pulling this
         * image or the sidecar image.
         */
        this.imagePullSecrets = [];
        /** Path to mount as the base path for executable code in the container.*/
        this.mountPath = "/src";
        /** Set the max time in miliseconds to wait for this job to complete.*/
        this.timeout = defaultTimeout;
        /** Fetch the source repo. Default: true*/
        this.useSource = true;
        /** If true, the job will be run in privileged mode.
         * This is necessary for Docker engines running inside the Job, for example.
         */
        this.privileged = false;
        /**
         * pod annotations for the job
         */
        this.annotations = {};
        /** _podName is set by the runtime. It is the name of the pod.*/
        this._podName = "";
        /** streamLogs controls whether logs from the job Pod will be streamed to output
         * this is similar to using `kubectl logs PODNAME -f`
         */
        this.streamLogs = false;
        if (!jobNameIsValid(name)) {
            throw new Error("job name must be lowercase letters, numbers, and '-', and must not start or end with '-', having max length " +
                Job.MAX_JOB_NAME_LENGTH);
        }
        this.name = name.toLocaleLowerCase();
        this.image = image || "";
        this.imageForcePull = imageForcePull;
        this.tasks = tasks || [];
        this.args = [];
        this.env = {};
        this.cache = new JobCache();
        this.storage = new JobStorage();
        this.docker = new JobDockerMount();
        this.host = new JobHost();
        this.resourceRequests = new JobResourceRequest();
        this.resourceLimits = new JobResourceLimit();
    }
    /** podName is the generated name of the pod.*/
    get podName() {
        return this._podName;
    }
}
Job.MAX_JOB_NAME_LENGTH = 36;
exports.Job = Job;
/**
 * jobNameIsValid checks the validity of a job's name.
 */
function jobNameIsValid(name) {
    return (name.length <= Job.MAX_JOB_NAME_LENGTH &&
        /^(([a-z0-9][-a-z0-9.]*)?[a-z0-9])+$/.test(name));
}
exports.jobNameIsValid = jobNameIsValid;
//# sourceMappingURL=job.js.map