const { events } = require("libacid")

events.on("exec", () => {
  console.log("==> handling an 'exec' event")
})
