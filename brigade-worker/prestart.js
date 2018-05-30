const process = require("process")
const fs = require("fs")

// Script locations in order of precedence.
const scripts = [
  process.env.BRIGADE_SCRIPT,

  // checked out in repo
  "/vcs/brigade.js",

  // mounted data from brigade.sh/build.Script
  "/etc/brigade/script",

  // mounted configmap named in brigade.sh/project.DefaultScriptName
  "/etc/brigade-default-script",
];

try {
  var data = loadScript()
  let wrapper = "const {overridingRequire} = require('./require');((require) => {" +
    data.toString() +
    "})(overridingRequire)"
  fs.writeFile("dist/brigade.js", wrapper, () => {
    console.log("prestart: src/brigade.js written")
  })
} catch(e) {
  console.log("prestart: no script override")
  process.exit(1)
}

// loadScript loads the first configured script it finds.
function loadScript() {
  for (let src of scripts) {
    if (fs.existsSync(src)) {
      var data = fs.readFileSync(src, 'utf8')
      if (data != "") {
        return data
      }
    }
  }
}
