const process = require("process")
const fs = require("fs")
const { execFileSync } = require("child_process")

const configFile = "/config.json";
const mountedConfigFile = "/etc/brigade/config";
const vcsConfigFile = "/vcs/brigade.json";

// Config file locations in order of precedence.
const configFiles = [
  // data mounted from event secret (e.g. brig run)
  mountedConfigFile,

  // checked out in repo
  vcsConfigFile,
];

function createConfig() {
  for (let src of configFiles) {
    if (fs.existsSync(src) && fs.readFileSync(src, "utf8") != "") {
      // Node's require will complain/fail if the file does not have a .json/.js extension
      // Here we create the appropriately named file using the contents from src
      fs.writeFileSync(configFile, fs.readFileSync(src, "utf8"));
      return;
    }
  }
}

if (require.main === module)  {
  addDeps()
}

function addDeps() {
  createConfig();
  if (!fs.existsSync(configFile)) {
    console.log("prestart: no dependencies file found")
    return
  }

  // Parse the config file
  // Currently, we only look for dependencies
  const deps = require(configFile).dependencies || {}

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
  configFile,
  mountedConfigFile,
  vcsConfigFile,
  createConfig,
  addDeps,
  buildPackageList,
  addYarn,
}
