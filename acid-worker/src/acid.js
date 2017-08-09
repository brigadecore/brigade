// This is the stub acid file.
const {events} = require("./libacid")

events.on("push", function(e, p){
  // TODO: Should this throw an exception so that the pod fails immediately?
  console.log("got event " + e.type + " for " + p.name)
  console.log("WARNING: No acid.js file was found.")
})
