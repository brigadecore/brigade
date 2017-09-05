const { events } = require("libacid")

events.on("exec", () => {
  console.log("==> handling an 'exec' event")
})

events.on("after", () => {
  console.log(" **** AFTER EVENT called")
})
