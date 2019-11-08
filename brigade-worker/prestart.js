const process = require("process")
const fs = require("fs")
const { execFileSync } = require("child_process")

const mountedDepsFile = "/etc/brigade/brigade.json";
const vcsDepsFile = "/vcs/brigade.json";
// Deps file locations in order of precedence.
const depsFiles = [
  // data mounted from event secret (e.g. brig run)
  mountedDepsFile,

  // checked out in repo
  vcsDepsFile,
];

function findDeps() {
  for (let src of depsFiles) {
    if (fs.existsSync(src) && fs.readFileSync(src, "utf8") != "") {
      return src;
    }
  }
  return "";
}

if (require.main === module)  {
  addDeps()
}

function addDeps() {
  const depsFile = findDeps();
  if (!depsFile) {
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
  mountedDepsFile,
  vcsDepsFile,
  addDeps,
  buildPackageList,
  addYarn,
}
