# Contributing Guide

## Signed commits

A DCO sign-off is a line placed at the end of a commit message containing a contributor's signature.
In adding this, the contributor certifies that they have the right to contribute the material.

Here are the steps to sign one's work:

Once the contributor certifies the DCO below (from [developercertificate.org](https://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
1 Letterman Drive
Suite D4700
San Francisco, CA, 94129

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

The contributor then just needs to add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.smith@example.com>

One's real name must be used (no pseudonyms or anonymous contributions).

The easiest way to do this is, assuming `user.name` and `user.email` are set via 
the git cli configuration (`git config`), is to sign commits automatically via `git commit -s`.

Finally, the `git log` information for a commit should show something like this:

```
Author: Joe Smith <joe.smith@example.com>
Date:   Thu Feb 2 11:41:15 2018 -0800

    Update README

    Signed-off-by: Joe Smith <joe.smith@example.com>
```

Notice the `Author` and `Signed-off-by` lines match. If they don't,
the PR will be rejected by the automated DCO check.