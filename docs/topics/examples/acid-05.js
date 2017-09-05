const { events, Job } = require("libacid")

events.on("exec", () => {
  var job = new Job("do-nothing", "alpine:3.4")

  job.run()
})
