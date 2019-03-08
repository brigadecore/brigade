const { events, Job } = require("brigadier")

events.on("exec", (e) => {
  var job = new Job("cacher", "alpine:3.4")
  job.cache.enabled = true

  job.tasks = [
    "echo " + e.buildID + " >> /mnt/brigade/cache/jobs.txt",
    "cat /mnt/brigade/cache/jobs.txt"
  ]

  job.run()
})
