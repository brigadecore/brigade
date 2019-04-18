const process = require("process")
const fs = require("fs")
const { execFileSync } = require("child_process")

//checked out in repo
const depsFile = "/vcs/brigade.json"

if (require.main === module)  {
  addDeps()
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
