---
title: Releasing Brigade
description: How to cut a new release of Brigade
aliases:
  - /releasing.md
  - /intro/releasing.md
  - /topics/releasing.md
---

# Releasing Brigade

Once the intended commit has been tested and we have confidence to cut a release,
we can follow these steps to release Brigade:

1. Issue a docs pull request with all `<current release>` strings updated to 
`<anticipated release>`, e.g. `1.2.0` becomes `1.2.1`.

1. Once this pull request is merged, create and push the git tag from the intended commit:

    ```console
    $ git tag v1.2.1
    $ git push origin v1.2.1
    ```

    The release pipeline located in our [brigade.js](../../brigade.js) then takes over
    and does the heavy lifting of building component images, pushing to designated
    image registries, building the `brig` cli binaries and finally creating the
    [GitHub release](https://github.com/brigadecore/brigade/releases).

1. Lastly, issue a pull request in [brigadecore/charts][charts]
bumping the `version` and `appVersion` values in both the Brigade
[chart](https://github.com/brigadecore/charts/blob/main/charts/brigade/Chart.yaml) and
the Brigade Project [chart](https://github.com/brigadecore/charts/blob/main/charts/brigade-project/Chart.yaml)
to match the current release value.  Once this pull request is merged, the
[brigade.js pipeline](https://github.com/brigadecore/charts/blob/main/brigade.js) will handle building
fresh chart artifacts and updating the chart index file.

[charts]: https://github.com/brigadecore/charts