---
title: Releasing Brigade
description: How to cut a new release of Brigade
aliases:
  - /releasing.md
---

# Releasing Brigade

Once the intended commit has been tested and we have confidence to cut a release,
we can follow these steps to release Brigade:

1. Issue a docs pull request with all `<current release>` strings updated to 
`<anticipated release>`, e.g. `2.0.0-alpha.5` becomes `2.0.0-beta.1`.

1. Once this pull request is merged, create and push the git tag from the intended commit:

    ```console
    $ git tag v2.0.0-beta.1
    $ git push origin v2.0.0-beta.1
    ```

    The release pipeline located in our [brigade.ts](../../.brigade/brigade.ts) then takes over
    and does the heavy lifting of building component images, pushing to designated
    image registries, publishing Helm charts, building the `brig` cli binaries and finally\
    creating the [GitHub release](https://github.com/brigadecore/brigade/releases).