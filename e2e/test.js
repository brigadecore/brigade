const { events } = require("brigadier")

events.on("exec", () => {
  console.log("==> handling an 'exec' event")
})