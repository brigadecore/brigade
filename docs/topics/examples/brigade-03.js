const { events } = require("brigadier")

events.on("exec", () => {
  console.log("==> handling an 'exec' event")
})

events.on("push", () => {
  console.log(" **** I'm a GitHub 'push' handler")
})
