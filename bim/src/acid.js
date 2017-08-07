// This is the stub acid file.
const {events, Job} = require("./libacid")

events.on("push", function(e, p){
  console.log("got event " + e.type + " for " + p.name)

  // Run a job and wait for it to return
  var j = new Job("foo", "alpine:3.4", ["sleep 10", "echo hello"])

  j.useSource = false
  j.timeout = 30000
  console.log(j)

  j.run().then(res => {
    console.log("finished job " + j.name)
    console.log(res.toString())
  }).catch(err => {
    console.log("error on job " + j.name)
    console.log(err)
  })

  /*

  // start two jobs and wait for them both to return
  var j1 = new Job("build-code")
  var j2 = new Job("build-docs")

  var g = new Group()
  g.add(j1)
  g.add(j2)

  let resultList = g.run()
  resultList.forEach((item) => {
    console.log(item.toString())
  })

  // start a job, and get a Promise for when it returns:
  j3 = new Job("last")
  var p = j3.background()

  // We can now use the promise:
  p.then(res => {
    console.log(res.toString())
  })
  */
})
