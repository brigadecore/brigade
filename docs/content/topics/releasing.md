---
title: Releasing Brigade 2
description: How to cut a new release of Brigade 2
section: topics
weight: 9
aliases:
  - /releasing.md
  - /intro/releasing.md
  - /topics/releasing.md
---

# Releasing Brigade 2

Releasing Brigade 2 is, generally speaking, easy, and mostly automated. Some
complications are introduced in the _rare_ event that the API version supported
by the server is changing.

This document breaks down the steps for cutting a release of Brigade 2 into:

1. Pre-release steps
2. Release steps
3. Post release steps

## Pre-Release

For the common case (API version supported by the server is not changing), only
the following files require updates. They should be updated to reflect the
version number we intend to assign to the release.

* `README.md`
* `docs/content/intro/quickstart.md`
* `docs/content/topics/operators/deploy.md`

### Submit a PR

All pre-release changes should be incorporated into a PR opened by a maintainer
and signed off upon by another maintainer.

## Release

When the PR containing pre-release changes have been merged and post-merge CI
has completed without error on the `main` branch, it is safe to cut the release
through application of a semver tag.

The _safest_ way to do this is through the GitHub UI since that helps mitigate
the possibility of accidentally tagging the wrong commit.

1. Go to
[https://github.com/brigadecore/brigade/releases/new](https://github.com/brigadecore/brigade/releases/new).
1. Apply the desired semver tag to the head of the `main` branch.
1. Use the same value as the title of the release.
1. Click __Publish release__.

Release automation should take everything from there.

To release the SDK for Go, a special tag has to be applied to the same commit as
the semver tag. Without this tag, Go's dependency management will be unable to
locate this release of the SDK for Go. For lack of a better alternative, this
really needs to be done via the `git` CLI.

1. Make sure your local `main` branch is up to date with respect to the remote
   `github.com/brigadecore/brigade` repository.
1. On the head of the `main` branch, run `git tag sdk/<semver>` where `<semver>`
   is equal to the semantic version applied to this release, _including the
   leading `v`_.

## Post-Release

Post-release, all examples having a `package.json` should be updated to use the
newly released `@brigadecore/brigadier` NPM package. This is not strictly
necessary because the worker component substitutes its own alternative version
of that package at runtime anyway, but these changes are still advised merely
for the sake of keeping examples tidy and up to date.

For every example having a `package.json` file:

1. Update `package.json` to reference the latest version of the
   `@brigadecore/brigadier` package (the one we just released).
1. If the example utilizes `npm` as a package manager, run `npm install` from
   the example's `.brigade/` directory.
1. If the example utilizes `yarn` as a package manager, run `yarn install` from
   the example's `.brigade/` directory.

### Submit a PR

All post-release changes should be incorporated into a PR opened by a maintainer
and signed off upon by another maintainer.
