const {events} = require("brigadier")

events.on("exec", function(e, project) {
  const e2 = {
    type: "next",
    provider: "exec-handler",
    buildID: e.buildID,
    workerID: e.workerID,
    cause: {event: e}
  }
  events.fire(e2, project)
})

events.on("next", (e) => {
  console.log(`fired ${e.type} caused by ${e.cause.event.type}`)
})
