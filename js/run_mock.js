
// This package mocks the run.js functions. It can replace run.js for testing.

function waitForJob(job) {
  return true
}

function run(job, e) {
  return job.name + "-" + Date.now() + "-" + e.commit.substring(0, 8);
}

exports.waitForJob = waitForJob
exports.run = run
