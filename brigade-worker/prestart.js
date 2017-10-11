const process = require("process")
const fs = require("fs")

const script = process.env.BRIGADE_SCRIPT || "/etc/brigade/script"

try {
  var data = fs.readFileSync(script, 'utf8')
  let wrapper = "const {whitelistRequire} = require('./require');((require) => {" +
    data.toString() +
    "})(whitelistRequire)"
  fs.writeFile("dist/src/brigade.js", wrapper, () => {
    console.log("prestart: src/brigade.js written")
  })
} catch(e) {
  console.log("prestart: no script override")
  process.exit(1)
}
