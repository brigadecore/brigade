const process = require("process")
const fs = require("fs")
const { execFileSync } = require("child_process")

// Script locations in order of precedence.
const scripts = [
  // manual override for debugging
  process.env.BRIGADE_SCRIPT,

  // data mounted from event secret (e.g. brig run)
  "/etc/brigade/script",

  // checked out in repo
  "/vcs/brigade.js",

  // data mounted from project.DefaultScript
  "/etc/brigade-project/defaultScript",

  // mounted configmap named in brigade.sh/project.DefaultScriptName
  "/etc/brigade-default-script/brigade.js"
];

//checked out in repo
const depsFile = "/vcs/brigade.json"

if (require.main === module)  {
  addDeps()

  try {
    var data = loadScript()
    let wrapper = "const {overridingRequire} = require('./require');((require) => {" +
      data.toString() +
      "})(overridingRequire)"
    fs.writeFileSync("dist/brigade.js", wrapper)
  } catch (e) {
    console.log("prestart: no script override")
    console.error(e)
    process.exit(1)
  }
}

// loadScript loads the first configured script it finds.
function loadScript() {
  for (let src of scripts) {
    if (fs.existsSync(src)) {
      var data = fs.readFileSync(src, 'utf8')
      if (data != "") {
        console.log(`prestart: loading script from ${ src }`)
        return data
      }
    }
  }
}

function addDeps() {
  if (!fs.existsSync(depsFile)) {
    console.log("prestart: no dependencies file found")
    return
  }
  const deps = require(depsFile).dependencies || {}

  const packages = buildPackageList(deps)
  if (packages.length == 0) {
    console.log("prestart: no dependencies to install")
    return
  }

  console.log(`prestart: installing ${packages.join(', ')}`)
  try {
    addYarn(packages)
  } catch (e)  {
    console.error(e)
    process.exit(1)
  }
}

function buildPackageList(deps) {
  if (!deps) {
    throw new Error("'deps' must not be null")
  }

  return Object.entries(deps).map(([dep, version]) => dep + "@" + version)
}

function addYarn(packages) {
  if (!packages || packages.length == 0) {
    throw new Error("'packages' must be an array with at least one item")
  }

  execFileSync("yarn", ["add", ...packages])
}

module.exports = {
  depsFile,
  addDeps,
  buildPackageList,
  addYarn,
}
