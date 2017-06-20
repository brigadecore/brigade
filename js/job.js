// The default image is stock ubuntu 16.04 + make and git.
var acidImage = "acid-ubuntu:latest"

// Prototype for Job.
function Job(name, tasks) {
  var my = this

  if (!exports._event) {
    throw "event not found"
  }

  // Name will become the prefix for the pod/configmap names.
  this.name = name;
  // Tasks is the list of tasks to run. They are executed in sequence inside of
  // a shell (/bin/sh).
  this.tasks = tasks;

  // A collection of name/value pairs of environment variables.
  this.env = {};

  // The image and an optional tag.
  this.image = acidImage;

  // A map of ENV_VAR names and Secret names. This will populate the environment
  // variable with the value found in the secret.
  // This will override a matching env var from the env map.
  this.secrets = {}

  // podName is set by run(), and contains the name of the pod created.
  this.podName

  // run sends this job to Kubernetes.
  this.run = function() {
    this.background(exports._event)
    this.wait()

    return this
  };

  this.background = function() {
    this.podName = run(this, exports._event);
  };

  // waitUntilDone is here for backwards compatibility, but does nothing.
  // DEPRECATED: Will be removed during Alpha
  this.waitUntilDone = function() {

    return this
  }

  // wait waits until a pod hits "Succeeded"
  //
  // wait() can be called on backgrounded objects.
  //
  // wait() is automatically called by this.run.
  //
  // If pod returns "Failed", this throws an exception.
  // If pod runs for more than 15 minutes (300 * 3-second intervals), throws a timeout exception.
  this.wait = function() {
    waitForJob(my)
  };
}

exports.Job = Job
