var { events, Job } = require("brigadier")

events.on("exec", () => {
  console.log("fire!")
})
