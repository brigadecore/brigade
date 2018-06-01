const process = require("process")
const fs = require("fs")
const exec = require("child-process-promise")

// Script locations in order of precedence.
const scripts = [
  // manual override for debugging
  process.env.BRIGADE_SCRIPT,

  // data mounted from event secret (e.g. brig run)
  "/etc/brigade/script",

  // checked out in repo
  "/vcs/brigade.js",

  // mounted configmap named in brigade.sh/project.DefaultScriptName
  "/etc/brigade-default-script",
];

//checked out in repo
const deps = "/vcs/brigade.json"

addDeps()

try {
  var data = loadScript()
  let wrapper = "const {overridingRequire} = require('./require');((require) => {" +
    data.toString() +
    "})(overridingRequire)"
  fs.writeFile("dist/brigade.js", wrapper, () => {
    console.log("prestart: src/brigade.js written")
  })
} catch (e) {
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

function addDeps() {
  if (fs.existsSync(deps)) {
    const p = require(deps)
    for (var dep in p.dependencies) {
      var d = dep + "@" + p.dependencies[dep];
      console.log("installing " + d)
      addYarn(d);
    }
  } else {
    console.log("prestart: no dependencies file found")
  }
}

function addYarn(arg) {
  return exec.exec(`yarn add ${arg}`, {})
    .catch(e => {
      console.error(e)
      process.exit(1)
    });
}
