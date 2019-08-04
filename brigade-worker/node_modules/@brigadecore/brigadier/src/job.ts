/**
 * Package job provides support for jobs.
 *
 * A Job idescribes a particular unit of a build. A Job returns a Result.
 * A JobRunner is an implementation of the runtime logic for a Job.
 */

/** */

import { V1EnvVarSource, V1VolumeMount, V1Volume } from "@kubernetes/client-node/dist/api";

/**
 * The default shell for the job.
 */
const defaultShell: string = "/bin/sh";
/**
 * defaultTimeout is the default timeout for a job (15 minutes)
 */
const defaultTimeout: number = 1000 * 60 * 15;
/**
 * The default image if `Job.image` is not set
 */
const brigadeImage: string = "debian:jessie-slim";

export const brigadeCachePath = "/mnt/brigade/cache";
export const brigadeStoragePath = "/mnt/brigade/share";
export const dockerSocketMountPath = "/var/run/docker.sock";
export const dockerSocketMountName = "docker-socket";

/**
 * JobRunner is capable of executing a job within a runtime.
 */
export interface JobRunner {
  // TODO: Should we add the constructor here?
  // Start runs a new job. It returns a JobRunner that can be waited upon.
  start(): Promise<JobRunner>;
  // Wait waits unti the job being run has reached a success or failure state.
  wait(): Promise<Result>;
}

/**
 * Result is the result of a particular Job.
 *
 * Every Result can be converted to a String with the `toString()` function. The
 * string is human-readable.
 */
export interface Result {
  toString(): string;
}

/**
 * Cache controls the job's cache.
 *
 * A cache is a small storage space that is shared between different instances
 * if the same job.
 *
 * Cache is just a plain filesystem, and as such comes with no guarantees about
 * consistency, etc. It should be treated as volatile.
 */
export class JobCache {
  /**
   * If enabled=true, a storage cache will be attached.
   */
  public enabled: boolean = false;
  /**
   * size is the amount of storage space assigned to the cache. The default is
   * 5Mi.
   * For sizing information, see https://github.com/kubernetes/community/blob/master/contributors/design-proposals/resources.md
   */
  public size: string = "5Mi";

  // EXPERIMENTAL: Allow script authors to change this location.
  // Before Brigade 0.15, this used a getter to prevent scripters from setting
  // this path directly.
  public path: string = brigadeCachePath;
}

/**
 * JobStorage configures build-wide storage preferences for this job.
 *
 * Changes to this object only impact the job, not the entire build.
 */
export class JobStorage {
  public enabled: boolean = false;

  // EXPERIMENTAL: Allow setting the path.
  // Prior to Brigade 0.15, this was read-only.
  public path: string = brigadeStoragePath;
}

/**
 * JobHost expresses expectations about the host a job will run on.
 */
export class JobHost {
  constructor() {
    this.nodeSelector = new Map<string, string>();
  }

  /**
   * os is the name of the OS upon which the job's container must run.
   *
   * This allows users to indicate that the container _must_ run on
   * "windows" or "linux" hosts. It is primarily useful in a "mixed node"
   * environment where the brigade.js will be run on a cluster that has more than
   * one OS
   */
  public os?: string;

  /**
   * name of the host to run on.
   *
   * If this is set, a job will ask to be run on this named host. Generally, this
   * should be used only if it is necessary to run the job on a particular host.
   * If not set, Brigade will let the scheduler decide, which is strongly preferred.
   *
   * Example usage: If you use a Kubernetes-ACI bridge, you may want to use this
   * to run jobs on the bridge.
   */
  public name?: string;

  /**
   * nodeSelector labels are used as selectors when choosing a node on which to run this job.
   */
  public nodeSelector: Map<string, string>;
}

/**
 * JobDockerMount enables or disables mounting the host's docker socket for a job.
 */
export class JobDockerMount {
  /**
   * enabled configues whether or not the job will mount the host's docker socket.
   */
  public enabled: boolean = false;
}

/**
 * JobResourceRequest represents request of the resources
 */
export class JobResourceRequest {
  /** cpu requests */
  public cpu?: string;
  /** memory requests */
  public memory?: string;
}

/**
 * JobResourceLimit represents limit of the resources
 */
export class JobResourceLimit {
  /** cpu limits */
  public cpu?: string;
  /** memory limits */
  public memory?: string;
}

/**
 * EXPERIMENTAL: allow mounting volumes to the job runner pod.
 * JobVolumeConfig represents a Kubernetes Volume and VolumeMount pair.
 * The Volume is mounted in the job pod at the path defined by the VolumeMount.
 * The name for the mount and volume must match. 
 * For more info, see https://kubernetes.io/docs/concepts/storage/volumes/
 * 
 * For a simple shared volume between all the containers of a job, use JobStorage.
 */
export class JobVolumeConfig {
  /** mounting config of the volume */
  public mount?: V1VolumeMount;
  /** volume that will be mounted at the mount location */
  public volume?: V1Volume;

  /**
   * 
   * @param m represents the volume mount
   * @param v represents the volume
   */
  constructor(m: V1VolumeMount, v: V1Volume) {
    this.mount = m;
    this.volume = v;
  }
}

/**
 * Job represents a single job, which is composed of several closely related sequential tasks.
 * Jobs must have names. Every job also has an associated image, which references
 * the Docker container to be run.
 * */
export abstract class Job {
  public static readonly MAX_JOB_NAME_LENGTH = 36;

  /** name of the job*/
  public name: string;
  /** shell that will be used by default in this job*/
  public shell: string = defaultShell;
  /** tasks is a list of tasks run inside of the shell*/
  public tasks: string[];
  /** args is a list of arguments that will be supplied to the container.*/
  public args: string[];
  /** env is the environment variables for the job*/
  public env: { [key: string]: string | V1EnvVarSource };
  /** image is the container image to be run*/
  public image: string = brigadeImage;
  /** imageForcePull defines the container image pull policy: Always if true or IfNotPresent if false */
  public imageForcePull: boolean = false;
  /**
   * imagePullSecrets names secrets that contain the credentials for pulling this
   * image or the sidecar image.
   */
  public imagePullSecrets: string[] = [];

  /** Path to mount as the base path for executable code in the container.*/
  public mountPath: string = "/src";

  /** Set the max time in miliseconds to wait for this job to complete.*/
  public timeout: number = defaultTimeout;

  /** Fetch the source repo. Default: true*/
  public useSource: boolean = true;

  /** If true, the job will be run in privileged mode.
   * This is necessary for Docker engines running inside the Job, for example.
   */
  public privileged: boolean = false;

  /** The account identity to be used when running this job.
   * This is an optional way to override the build-wide service account. If it is
   * not specified, the main worker service account will be used.
   *
   * Different Brigade worker implementations may choose to honor or ignore this
   * for security or configurability reasons.
   *
   * See https://github.com/brigadecore/brigade/issues/251
   */
  public serviceAccount?: string;

  /** Set the resource requests for the containers */
  public resourceRequests: JobResourceRequest;

  /** Set the resource limits for the containers */
  public resourceLimits: JobResourceLimit;

  /**
   * host expresses expectations about the host the job will run on.
   */
  public host: JobHost;

  /**
   * cache controls per-Job caching preferences.
   */
  public cache: JobCache;
  /**
   * storage controls this job's preferences on the build-wide storage.
   */
  public storage: JobStorage;

  /**
   * volume configuration preferences for current job.
   */
  public volumeConfig: JobVolumeConfig[];
  /**
   * docker controls the job's preferences on mounting the host's docker daemon.
   */
  public docker: JobDockerMount;

  /**
   * pod annotations for the job
   */
  public annotations: { [key: string]: string } = {};

  /** _podName is set by the runtime. It is the name of the pod.*/
  protected _podName: string = "";

  /** podName is the generated name of the pod.*/
  get podName(): string {
    return this._podName;
  }

  /** streamLogs controls whether logs from the job Pod will be streamed to output
   * this is similar to using `kubectl logs PODNAME -f`
   */
  public streamLogs: boolean = false;

  /** Create a new Job
   * name is the name of the job.
   * image is the container image to use
   * tasks is a list of commands to run.
   */
  constructor(
    name: string,
    image?: string,
    tasks?: string[],
    imageForcePull: boolean = false
  ) {
    if (!jobNameIsValid(name)) {
      throw new Error(
        "job name must be lowercase letters, numbers, and '-', and must not start or end with '-', having max length " +
        Job.MAX_JOB_NAME_LENGTH
      );
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
    this.volumeConfig = [];
  }

  /** run executes the job and then */
  public abstract run(): Promise<Result>;

  /** logs retrieves the logs (so far) from the job run */
  public abstract logs(): Promise<string>;
}

/**
 * jobNameIsValid checks the validity of a job's name.
 */
export function jobNameIsValid(name: string): boolean {
  return (
    name.length <= Job.MAX_JOB_NAME_LENGTH &&
    /^(([a-z0-9][-a-z0-9.]*)?[a-z0-9])+$/.test(name)
  );
}
