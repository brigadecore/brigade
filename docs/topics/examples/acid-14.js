const { events, Job } = require("libacid")

events.on("exec", (e, p) => {
  var echo = new Job("echo", "alpine:3.4")
  echo.tasks = [
    "echo Project " + p.name,
    "echo Event $EVENT_NAME"
  ]

  echo.env = {
    "EVENT_NAME": e.type
  }

  echo.run()
})
