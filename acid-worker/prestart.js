const process = require("process")
const {Buffer} = require("buffer")
const fs = require("fs")

var script = process.env.ACID_SCRIPT
if (!script || script == "") {
  console.log("prestart: no script override")
  process.exit(0)
}

var buf = Buffer.from(script, "base64")
fs.writeFile("src/acid.js", buf, () => {
  console.log("prestart: src/acid.js written")
})
