const { events, Job, Group } = require("brigadier")

events.on("exec", () => {
  var hello = new Job("hello", "alpine:3.4", ["echo hello"])
  var goodbye = new Job("goodbye", "alpine:3.4", ["echo goodbye"])

  var helloAgain = new Job("hello-again", "alpine:3.4", ["echo hello again"])
  var goodbyeAgain = new Job("bye-again", "alpine:3.4", ["echo bye again"])


  var first = new Group()
  first.add(hello)
  first.add(goodbye)

  var second = new Group()
  second.add(helloAgain)
  second.add(goodbyeAgain)

  first.runAll().then( () => second.runAll() )
})
