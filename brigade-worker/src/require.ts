const path = require("path");

const pkgOverrides: Pkg[] = [
  { name: "brigade", override: "./brigadier" },
  { name: "brigadier", override: "./brigadier" },
  { name: "@brigadecore/brigadier", override: "./brigadier" }
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
  return require(getOverriddenPackage(pkg));
}

export function getOverriddenPackage(pkg: string): string {
  for (let p of pkgOverrides) {
    if (p.name == pkg) {
      return p.override;
    }

    // we want to intercept the loading of relative modules in the repo
    // and and override them with the absolute path inside the worker pod
    if ((!path.isAbsolute(pkg)) && (pkg.includes("./"))) {
      return pkg.substr(0, pkg.indexOf("./")) + "/vcs" + pkg.substr(pkg.indexOf("./") + 1);
    }
  }
  return pkg;
}
