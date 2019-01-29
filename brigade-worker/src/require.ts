const path = require("path");

const pkgOverrides: Pkg[] = [
  { name: "brigade", override: "./brigadier" },
  { name: "brigadier", override: "./brigadier" },
  { name: "@azure/brigadier", override: "./brigadier" }
];

class Pkg {
  /**
   * The name of the package to be whitelisted.
   */
  name: string;
  /**
   * An optional library that will be loaded instead of the named library.
   */
  override?: string;
}

export function overridingRequire(pkg: string): any {
  for (let p of pkgOverrides) {
    if (p.name == pkg) {
      return require(p.override);
    }

    // we want to intercept the loading of relative modules in the repo
    // and and override with the absolute path inside the worker pod
    if ((!path.isAbsolute(pkg)) && (pkg.includes("./"))) {
      return require(pkg.substr(0, pkg.indexOf("./")) + "/vcs" + pkg.substr(pkg.indexOf("./") + 1))
    }
  }
  return require(pkg);
}
