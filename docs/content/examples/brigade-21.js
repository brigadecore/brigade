const {events} = require("brigadier")

events.on("exec", (e, project) => {
  // This is only registered when 'exec' is called.
  events.on("next", () => {
    console.log("fired 'next' event")
  })
  events.emit("next", e, project)
})

events.on("exec2", (e, project) => {
  events.emit("next", e, project)
})

