---
title: The brig CLI
description: Using the brig CLI
weight: 4
aliases:
  - /brig
  - /topics/brig.md
  - /topics/project-developers/brig.md
---

The `brig` CLI provides access to the full repertoire of supported user
interactions in Brigade, whether it's logging into Brigade with `brig login`,
bootstrapping a new Brigade project with `brig init`, creating events with
`brig event create` -- the list goes on.

In this doc, we'll go over how to [install brig] and then give a brief overview
of the [suite of commands] that brig provides.

[install brig]: #install-brig
[suite of commands]: #suite-of-commands

## Install brig

In general, `brig` can be installed by downloading the appropriate pre-built
binary from our [releases page](https://github.com/brigadecore/brigade/releases)
to a directory on your machine that is included in your `PATH` environment
variable. On some systems, it is even easier than this.

You can also build brig from source; see the [Developers] guide for more info.

[Developers]: /topics/developers

**linux**

```shell
curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-linux-amd64
chmod +x /usr/local/bin/brig
```

**macos**

The popular [Homebrew](https://brew.sh/) package manager provides the most
convenient method of installing the Brigade CLI on a Mac:

```shell
$ brew install brigade-cli
```

Alternatively, you can install manually by directly downloading a pre-built
binary:

```shell
$ curl -Lo /usr/local/bin/brig https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-darwin-amd64
$ chmod +x /usr/local/bin/brig
```

**windows**

```powershell
> mkdir -force $env:USERPROFILE\bin
> (New-Object Net.WebClient).DownloadFile("https://github.com/brigadecore/brigade/releases/download/v2.3.1/brig-windows-amd64.exe", "$ENV:USERPROFILE\bin\brig.exe")
> $env:PATH+=";$env:USERPROFILE\bin"
```

The script above downloads brig.exe and adds it to your PATH for the current
session. Add the following line to your [PowerShell Profile] to make the change
permanent.

```powershell
> $env:PATH+=";$env:USERPROFILE\bin"
```

[releases]: https://github.com/brigadecore/brigade/releases
[PowerShell Profile]: https://www.howtogeek.com/126469/how-to-create-a-powershell-profile/

## Suite of Commands

To view the full suite of commands that brig supports, simply type `brig` in
your console. You should see the commands available under `COMMANDS`. These
include:

  * `event`: Create and manage Brigade [Events]
  * `init`: Bootstrap a new Brigade [Project]
  * `login`: Log in to Brigade
  * `logout`: Log out of Brigade
  * `project`: Create and manage Brigade [Projects]
  * `role`: Grant, revoke and list system roles for [users] or [service accounts]
  * `service-account`: Create and manage [service accounts]
  * `users`: Manage authenticated [users]

Type any of these commands to get a help menu and start digging deeper into the
full selection of functionality that each provides. For example:

```plain
 $ brig event

NAME:
   Brigade event - Manage events

USAGE:
   Brigade event command [command options] [arguments...]

COMMANDS:
   cancel           Cancel a single event without deleting it
   cancel-many, cm  Cancel multiple events without deleting them
   clone            Clone an existing event
   create           Create a new event
   delete           Delete a single event
   delete-many, dm  Delete multiple events
   get              Retrieve an event
   list, ls         List events
   retry            Retry an event
   log, logs        View worker or job logs
   help, h          Shows a list of commands or help for one command

OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

[Events]: /topics/project-developers/events
[Project]: /topics/project-developers/projects
[Projects]: /topics/project-developers/projects
[users]: /topics/administrators/authorization
[service accounts]: /topics/administrators/authorization
