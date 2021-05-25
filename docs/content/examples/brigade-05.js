const { events, Job } = require("brigadier")

events.on("exec", () => {
  var job = new Job("do-nothing", "alpine:3.4")

  job.run()
})
