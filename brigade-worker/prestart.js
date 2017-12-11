const process = require("process")
const fs = require("fs")

const script = process.env.BRIGADE_SCRIPT || "/etc/brigade/script"
const vcsScript = process.env.BRIGADE_VCS_SCRIPT || "/vcs/brigade.js"

try {
  var data = loadScript(script)
  let wrapper = "const {whitelistRequire} = require('./require');((require) => {" +
    data.toString() +
    "})(whitelistRequire)"
  fs.writeFile("dist/brigade.js", wrapper, () => {
    console.log("prestart: src/brigade.js written")
  })
} catch(e) {
  console.log("prestart: no script override")
  process.exit(1)
}

// loadScript tries to load the configured script. But if it can't, it falls
// back to the VCS copy of the script.
function loadScript(script) {
  // This happens if the secret volume is mis-mounted, which should never happen.
  if (!fs.existsSync(script)) {
    console.log("prestart: no script found. Falling back to VCS script")
    return fs.readFileSync(vcsScript, 'utf8')
  }
  var data = fs.readFileSync(script, 'utf8')
  if (data == "") {
    // This happens if no file was submitted by the consumer.
    console.log("prestart: empty script found. Falling back to VCS script")
    return fs.readFileSync(vcsScript, 'utf8')
  }
  return data
}
