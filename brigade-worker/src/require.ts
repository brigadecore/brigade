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
  }
  return require(pkg);
}
