const {events} = require("brigadier")

events.on("exec", () => {
  console.log("first")
})

events.on("exec", () => {
  console.log("second")
})

