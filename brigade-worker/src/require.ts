let whitelist: Pkg[] = [
  {name: "brigade", override: "./brigadier"},
  {name: "brigadier", override: "./brigadier"},
  // Libraries that do not allow file or network I/O
  {name: "url"},
  {name: "util"},
  {name: "querystring"},
  {name: "path"},
  {name: "events"},
  // It's not clear if some of crypto's setter functions might have unintended
  // consequences on other uses of the crypto library. See crypto.setEngine().
  // If this turns out to be a low risk or insignificant, we should enable the
  // library.
  // {name: "crypto"},
  // buffer.allocUnsafe could potentially allow other parts of the program to be
  // accessed. This could presumably include something like a Kubernetes token.
  // The present security model does not consider this to be a critical issue,
  // since this information is obtainable via other pods as well. So we leave
  // the buffer module. In the future, it might be good to mask that function.
  {name: "buffer"},
  {name: "assert"}
]

class Pkg {
  /**
   * The name of the package to be whitelisted.
   */
  name: string
  /**
   * An optional library that will be loaded instead of the named library.
   */
  override?: string
}

export function whitelistRequire(pkg: string): any {
  for (let p of whitelist) {
    if( p.name == pkg ) {
      if (p.override) {
        return require(p.override)
      }
      return require(p.name)
    }
  }
  throw "package not allowed: " + pkg
}
