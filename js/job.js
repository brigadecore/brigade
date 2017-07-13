// The default image is stock ubuntu 16.04 + make and git.
var acidImage = "acid-ubuntu:latest"

// The default terminal emulator that job tasks will be executed under.
var defaultTerminal = "/bin/sh"

// Prototype for Job.
function Job(name, tasks) {
  var my = this

  if (!exports._event) {
    throw "event not found"
  }

  // Name will become the prefix for the pod/configmap names.
  this.name = name;

  // Shell is the teminal emulator which tasks will run under.
  this.shell = defaultTerminal;
  
  // Tasks is the list of tasks to run. They are executed in sequence inside of
  // a shell (/bin/sh).
  this.tasks = [];
  if (tasks) {
    this.tasks = tasks
  }

  // A collection of name/value pairs of environment variables.
  this.env = {};

  // The image and an optional tag.
  this.image = acidImage;

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
    waitForJob(my, exports._event)
  };
}

exports.Job = Job
