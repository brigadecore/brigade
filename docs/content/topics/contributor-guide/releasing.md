---
title: Releasing Brigade
description: How to cut a new release of Brigade
section: contributor-guide
weight: 4
---

Releasing Brigade is easy and mostly automated.

This section exists primarily for the benefit of project maintainers and
outlines, in brief, the Brigade release process.

> ⚠️&nbsp;&nbsp;These steps are also generally applicable to other projects
> owned by the [@brigadecore](https://github.com/brigadecore) GitHub org.

## Releasing Server-Side Components

### Pre-Release

To prepare for a release, a project maintainer must open a PR containing
"pre-release version bumps." Such a PR should update all references to the
version number of the latest Brigade release so that they instead reflect the
version number of the _upcoming_ Brigade release.

> ⚠️&nbsp;&nbsp;These are primarily, but not exclusively documentation updates.

The PR should be reviewed and signed off upon by another project maintainer.

Merging these changes _prior_ to tagging a new release with a semantic version
number ensures that the commit referenced by the new tag contains documentation
that correctly references that same release. Since documentation is continuously
deployed, it is also crucial that such changes _not_ be merged too far in
advance of a planned release, as the result will be that documentation will
reference a version of the software that does not yet exist, and this can
confuse new users.

> ⚠️&nbsp;&nbsp;This entire step can optionally be skipped when planning a
> "pre-release," such as a release candidate if the maintainers do not desire
> for live documentation to be updated such that the release candidate is
> presented as the latest version of Brigade.

### Creating the Release

Brigade's automated release process is triggered by the creation of a new
release (or pre-release) in GitHub, which must also reference a tag that adheres
to semantic versioning practices. It is insufficient to _only_ apply a tag. The
creation of the GitHub release is the actual trigger for Brigade's release
automation.

To create a release:

1. Validate that the CI process has completed successfully on the `main` branch
   after the pre-release version bumps have been merged.

1. Browse to
[https://github.com/brigadecore/brigade/releases/new](https://github.com/brigadecore/brigade/releases/new).

1. Click __Choose a tag__ and type the semantic version number of the release.
   This tag does not exist yet, so click the button that will appear that says
   __Create new tag on publish__. 

1. Use the semantic version number as the release title.

1. If applicable check the box specifying __This is a pre-release__. Until such
   time that work a major revision of Brigade (v3) begins (and there are
   currently no plans for this), this is only applicable to release candidates.

1. Click __Publish release__.

Automated processes will complete the release.

### Automated Release Process

The automated release process does the following:

1. Builds amd64 _and_ arm64-based Docker images of all Brigade server-side
   components, cryptographically signs those images, and pushes them to each
   component's canonical OCI repository. (These repositories are hosted on
   Docker Hub.)

1. Generates an SBOM (software bill of materials) for each image and publishes
   it to the corresponding GitHub release page.

1. Publishes a Helm chart to a canonical OCI repository. (This repository is
   hosted on ghcr.io)

1. Builds the `brig` CLI for a variety of OSes and CPU architectures and
   publishes the pre-built binaries to the GitHub release page.

1. Publishes Brigadier (the library used for writing Brigade scripts) to
   [npmjs.com](https://www.npmjs.com/). This step is last because it is the only
   one that is not strictly idempotent.

### Post-Release

Following a release, it is optional, but recommended to open a second PR that
updates all example scripts which make use of Brigadier such that they use the
latest (just released) version.

## Releasing the Brigade SDK for Go

Source for the Brigade SDK for Go is housed in the same repository as the
Brigade server side components. This was a deliberate choice so that features
requiring enhancements to both the SDK and the API server would not require
coordination across multiple repositories.

Despite its source being housed in the same repository, Brigade SDK for Go is
versioned independently of Brigade's server-side components.

To cut a release of the SDK, all that is required is for the appropriate commit
in the canonical repository to be tagged with a semantic version number. There
is no automation involved because the existence of the tag is all that is
required for Go's module system to be able to locate a given version of the SDK.