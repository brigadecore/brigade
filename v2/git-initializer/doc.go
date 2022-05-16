package main

// This package implements a "hybrid" approach to interacting with remote git
// repositories to acquire source code required by a Brigade Worker or Job. The
// "hybrid" approach does whatever it can using the (pure Go) go-git library,
// but falls back on exec'ing git CLI commands when necessary. This is necessary
// because go-git cannot fetch from remotes hosted by certain git providers --
// namely Azure DevOps, but possibly others.
//
// In an ideal world, we'd have a library that addressed all out needs. libgit2
// (via Go bindings provided by git2go) comes close to meeting our needs, but
// cannot use refspecs that use a sha. I (krancour) expect this to be fixed
// within a few months of this writing (May 2022), and we can possibly revisit
// this component at that time.
