const { events, Job } = require("libacid")

events.on("exec", (e) => {
  var job = new Job("cacher", "alpine:3.4")
  job.cache.enabled = true

  job.tasks = [
    "echo " + e.buildID + " >> /mnt/acid/cache/jobs.txt",
    "cat /mnt/acid/cache/jobs.txt"
  ]

  job.run()
})
