const {events} = require("brigadier")

events.on("exec", function(e, project) {
  events.emit("next", e, project)
})

events.on("next", () => {
  console.log("fired 'next' event")
})
