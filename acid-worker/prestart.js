const process = require("process")
const fs = require("fs")

const script = process.env.ACID_SCRIPT || "/etc/acid/script"

try {
  var data = fs.readFileSync(script, 'utf8')
  let wrapper = "const la = require(\"./libacid\");((events, Job, Group, require) => {" +
    data.toString() +
    "})(la.events, la.Job, la.Group, () => { throw 'require is disabled' })"
  fs.writeFile("src/acid.js", wrapper, () => {
    console.log("prestart: src/acid.js written")
  })
} catch(e) {
  console.log("prestart: no script override")
  process.exit(1)
}
