// WaitGroup waits for multiple jobs to finish. It will throw an error
// as soon as a job reports an error.
function WaitGroup() {
  this.jobs = []

  // add adds a new job to the waitgroup
  this.add = function(job) {
    this.jobs.push(job)
  }

  // run runs every job in the group, and then waits for them to complete.
  this.run = function() {
    this.jobs.forEach(function (j) {
      j.background()
    })
    this.wait()
  }

  // wait waits until jobs are complete. Note that this does not run the jobs. They
  // must be started externally. (See WaitGroup.run or Job.background)
  this.wait = function() {
    this.jobs.forEach(function (j) {
      j.wait()
    })
  }
}

exports.WaitGroup = WaitGroup
