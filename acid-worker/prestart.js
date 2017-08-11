const process = require("process")
const {Buffer} = require("buffer")
const fs = require("fs")

var script = process.env.ACID_SCRIPT
if (!script || script == "") {
  console.log("prestart: no script override")
  process.exit(0)
}


var buf = Buffer.from(script, "base64")
let wrapper = "const la = require(\"./libacid\");((events, Job, Group, require) => {" +
  buf.toString("utf8") +
  "})(la.events, la.Job, la.Group, () => { throw 'require is disabled' })"
fs.writeFile("src/acid.js", wrapper, () => {
  console.log("prestart: src/acid.js written")
})
