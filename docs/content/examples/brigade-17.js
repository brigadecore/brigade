const { events, Job } = require("brigadier")

events.on("exec", (e, p) => {
  var one = new Job("one", "alpine:3.4")
  var two = new Job("two", "alpine:3.4")

  one.tasks = ["echo world"]
  one.run().then( result => {
    two.tasks = ["echo hello " + result.toString()]
    two.run().then( result2 => {
      console.log(result2.toString())
    })
  })
})
