---
title: Signing Commits
description: How to sign your commits
section: contributor-guide
weight: 2
---

All commits merged into Brigade's `main` branch MUST bear _two_ types of
signatures from the contributor who authored them -- one being necessary per
[CNCF](https://www.cncf.io/), which governs the Brigade project, and the other
necessary for assurances of every commit's origin.

> ⚠️&nbsp;&nbsp;While this document discusses the signature requirements for
> commits made to Brigade, these requirements are equally applicable to all
> other repositories owned by the [@brigadecore](https://github.com/brigadecore)
> GitHub org.

> ⚠️&nbsp;&nbsp;If you prefer learning through video, check out the
> [video adaptation](https://www.youtube.com/watch?v=uHkQzxzciLA) of this guide
> on our YouTube channel.

## DCO Sign-Off

A DCO (Developer Certificate of Origin) sign-off is a line placed at the end of
a commit message containing a contributor's "signature." In adding this, the
contributor certifies that they have the right to contribute the material in
question.

Here are the steps to sign your work:

1. Verify the contribution in your commit complies with the
   [terms of the DCO](https://developercertificate.org/).

1. Add a line like the following to your commit message:

   ```
   Signed-off-by: Joe Smith <joe.smith@example.com>
   ```

   You MUST use your legal name -- handles or other pseudonyms are not
   permitted.

   While you could manually add DCO sign-off to every commit, there is an easier
   way:

    1. Configure your git client appropriately. This is one-time setup.

       ```shell
       $ git config user.name <legal name>
       $ git config user.email <email address you use for GitHub>
       ```

       If you work on multiple projects that require a DCO sign-off, you can
       configure your git client to use these settings globally instead of only
       for Brigade:

       ```shell
       $ git config --global user.name <legal name>
       $ git config --global user.email <email address you use for GitHub>
       ```

    1. Use the `--signoff` or `-s` (lowercase) flag when making each commit.
       For example:

       ```shell
       $ git commit --message "<commit message>" --signoff
       ```

       If you ever make a commit and forget to use the `--signoff` flag, you
       can amend your commit with this information before pushing:

       ```shell
       $ git commit --amend --signoff
       ```

    1. You can verify the above worked as expected using `git log`. Your latest
       commit should look similar to this one:

       ```
       Author: Joe Smith <joe.smith@example.com>
       Date:   Thu Feb 2 11:41:15 2018 -0800

           Update README

           Signed-off-by: Joe Smith <joe.smith@example.com>
       ```

       Notice the `Author` and `Signed-off-by` lines match. If they do not, the
       PR will be rejected by the automated DCO check.

### GPG Signature

While the DCO sign-off asserts a contributor's right to make their contribution,
the GPG signature is required to cryptographically offer a stronger assurance of
the contributor's identity.

Since the rationale for this requirement may be non-obvious, a brief
justification may be in order. Strong assurances of the identity of every
commit's author may, after all, seem superfluous since pushing commits to GitHub
already requires authentication. This would be true if not for the fact that
authenticated users can _also_ push commits authored by others. This can occur,
for instance, in a scenario where multiple contributors have collaborated on a
PR. Requiring every commit to be signed individually by an author known to
GitHub ensures that the progeny of every single commit is known and traceable to
a GitHub user.

Here is a summary of the steps required to sign your work. Most of this is
one-time setup.

[More extensive documentation](https://docs.github.com/en/github/authenticating-to-github/managing-commit-signature-verification)
of this subject is available from GitHub.

> ⚠️&nbsp;&nbsp;The steps below cover a minimally-viable setup for
> cryptographically signing commits. User who are very serious about guarding
> their digital identity may wish to research this topic further and consider
> measures up to and including generating keys on an air-gapped machine, but
> these precautions are well beyond the scope of this documentation.

1. If you do not already have them, download and install
   [the GPG command line tools](https://www.gnupg.org/download/) for your
   operating system. Note that these tools may also be available as system
   packages.

1. If `gpg` was already installed, list available keys:

   ```shell
   $ gpg --list-secret-keys --keyid-format LONG
   ```

   If you wish to use one of these keys, skip the next step.

1. For `gpg` versions 2.1.17 or greater, use the following command and then
   follow the prompts. You should generate an __RSA__ key with a length of at
   __4096__ bits.

   ```shell
   $ gpg --full-generate-key
   ```

   If your `gpg` version is less than 2.1.17, use the following command instead:

   ```shell
   $ gpg --default-new-key-algo rsa4096 --gen-key
   ```

   Now that your new key has been generated, re-list available keys as in the
   previous step:

   ```shell
   $ gpg --list-secret-keys --keyid-format LONG
   ```

1. Identify the ID of the key you wish to use for cryptographically signing your
   commits. _Copy this value._
   
   In the example below, the key ID is `3AA5C34371567BD2` (found on the line
   beginning with `sec`):

   ```shell
   $ gpg --list-secret-keys --keyid-format LONG
   /Users/hubot/.gnupg/secring.gpg
   ------------------------------------
   sec   4096R/3AA5C34371567BD2 2016-03-10 [expires: 2017-03-10]
   uid                          Hubot 
   ssb   4096R/42B317FD4BA89E7A 2016-03-10
   ```

1. Configure your git client to use the desired key:

   ```shell
   $ git config user.signingkey <key id>
   ```

   If you work on multiple projects that require cryptographically signed
   commits, you can configure your git client to use this setting globally
   instead of only for Brigade repositories:

   ```shell
   $ git config --global user.signingkey <key id>
   ```

1. Next associate this new key with your GitHub account.

    1. Export the public half of the key:

       ```shell
       $ gpg --armor --export <key id>
       ```

    1. Copy the public key, beginning with
       `-----BEGIN PGP PUBLIC KEY BLOCK-----` and ending with
       `-----END PGP PUBLIC KEY BLOCK-----`.
    
    1. Visit [https://github.com/settings/keys](https://github.com/settings/keys)
       and click __New GPG key__ and follow the prompts.

1. With all setup, up to this point, being one-time setup, all that remains is
   to cryptographically sign every commit you contribute to the Brigade project.
   Automation will reject pull requests containing any commits
   that are _not_ cryptographically signed.
   
   Cryptographically signing your commits can be accomplished with similar ease
   to adding the DCO signature. Simply use the `--gpg-sign` or `-S` (uppercase)
   flag:

   ```shell
   $ git commit --message "<commit message>" --gpg-sign
   ```

   If you ever make a commit and forget to use the `--gpg-sign` flag, you
   can amend your commit with this information before pushing:

   ```shell
   $ git commit --amend --gpg-sign
   ```

   Recalling that DCO sign-off is _also_ required, the full command to satisfy
   all signing requirements is:

   ```shell
   $ git commit --message "<commit message>" --signoff --gpg-sign
   ```

   Or more succinctly:

   ```shell
   $ git commit --message "<commit message>" -s -S
   ```

   Lastly, if you wish to cryptographically sign _all_ commits for a Brigade
   repository, you may spare yourself from having to remember the `-S` flag
   by signing all commits by default:
   
   ```shell
   $ git config commit.gpgSign true
   ```
   
   Or to enable this globally:
   
   ```shell
   $ git config --global commit.gpgSign true
   ```

   > ⚠️&nbsp;&nbsp;No similar option exists for automating DCO sign-off, since
   > DCO sign-off requires an explicit, _per-commit_ attestation that the
   > contributor has a right to contribute the material in question.
