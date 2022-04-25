---
title: Code Review Guide
description: The Brigade guide to reviewing PRs
section: contributor-guide
weight: 6
---

This document outlines the Brigade code review process that is applied to all
contributions (pull requests) and covers things like who can review PRs, when
and _how_.

> ⚠️&nbsp;&nbsp;Throughout this document, "project" (with a lowercase "p") is
> used to denote individual software projects or source code repositories that
> are related in some way to Brigade and owned by the
> [@brigadecore](https://github.com/brigadecore) GitHub org. In contrast,
> "Project" (with an uppercase "P") is used to denote the entire breadth of
> Brigade as a "program" encompassing Brigade itself and _all_ related software
> projects or source code repositories owned by the __@brigadecore__ org.

## Reviewers

For the Brigade Project, __Reviewer__ is not a formal role in the way that a
__Project Maintainer__ or __Core Maintainer__ is. Here, a Reviewer is _anyone_
who reviews a PR. This is to say, feedback on any change is always welcome from
_any_ Community Participant who adheres to the
[CNCF Code of Conduct](https://github.com/cncf/foundation/blob/master/code-of-conduct.md).

Despite reviews and comments being welcome from all Community Participants,
those reviews and comments carry varying degrees of weight under different
circumstances.

### PRs Opened by Contributors

PRs opened by Contributors who are not also a Core Maintainer or a Project
Maintainer of the project in question must be reviewed and approved _by_ a Core
Maintainer or a Project Maintainer. Feedback is welcome from all Community
Participants, but ultimate approval or rejection of the PR is at the discretion
of a Core Maintainer or Project Maintainer.

### PRs Opened by Maintainers

PRs opened by Project Maintainers and Core Maintainers fall into two categories:

* __Ordinary PRs:__ These require a thorough review from a peer Core Maintainer
  or Project Maintainer, the same as if they'd been opened by any other
  Contributor. Feedback is welcome from all Community Participants, but ultimate
  approval or rejection of the PR is at the discretion of a Core Maintainer or
  Project Maintainer _who is not the author of the PR_.

* __Chore PRs:__ To prevent the relatively small number of Core Maintainers and
  Project Maintainers from becoming review bottlenecks, a Core Maintainer or
  Project Maintainer (if opening a PR for their own project) may use their own
  discretion to classify their PR as a "chore" by applying the "chore" label to
  the PR. This indicates that the PR is (subjectively) small in scope and is
  likely be understood by any Community Participant -- even those _without_ deep
  technical knowledge of Brigade. With this designation, _any_ Community
  Participant may review and approve the PR.

## When to Review a PR

If a PR is in draft form, review/comments are permissible, but assume that the
author is not explicitly _requesting_ any sort of review or approval.
Contributors often open draft PRs for "safe keeping" or just to create some
visibility into their in-progress work. Assume the author of a draft _knows_
more work is required to achieve a review-worthy state. With this being the
case, reviewing draft PRs is not always the best use of one's time.

## How to Review a PR

Some important objectives of a PR review are to:

* Prevent any Contributor (Core Maintainers included) from making changes
  unilaterally.
* Ensure changes are well-thought-out and logically correct.
* Ensure a high level of code quality.
* Ensure changes are maintainable over the long term and do not introduce a
  maintenance burden.
* Ensure changes are well-tested.
* Ensure changes are well-documented -- both internally and in user-facing
  documentation.
* Avoid the introduction of new bugs or regressions.
* Discourage the introduction of breaking changes (except when a major release
  is planned).
* Prevent the introduction of malicious code.

Any Community Participant may comment on a PR to seek clarification on or
provide feedback on the proposed changes, so long as the comments are
constructive and respectful.

When reviewing PRs, it is often best to avoid
["bikeshedding"](https://en.wiktionary.org/wiki/bikeshedding)
on trivial matters such as coding style and use of whitespace. For the most
part, code style violations will be detected by lint checks that run as part of
the CI process. For other small, "nitpicky" bits of feedback, it is often
helpful to acknowledge your comment is a "nit." By doing so, you are leaving it
to the discretion of the PR author to act on that feedback or not. To summarize:
Effective reviews avoid _overwhelming_ PR authors with trivialities.

Any Community Participant may approve a PR, but this expression of approval may
or may not be considered binding. Refer to the previous section for
clarification on whose approval is binding under various circumstances.

Rejecting or closing a PR is solely within the purview of Core Maintainers and
applicable Project Maintainers.

## How to Merge a PR

As a courtesy, anyone having sufficient permissions to merge their own PR should
be permitted to merge it themselves after it has been approved. Admittedly, this
is for no other reason than the deep satisfaction that some Contributors obtain
through clicking the merge button themselves. Primarily, it will be Project
Maintainers and Core Maintainers who are able to do this.

For any PR authored by a Contributor _without_ sufficient permissions to merge
their own PR, the PR should be merged by a Project Maintainer or Core Maintainer
-- preferably one who has been formally assigned responsibility for reviewing
the PR and assisting its author with seeing their changes through to completion.

For PRs containing many commits, especially if the commit history is messy, the
"Squash and merge" option should be utilized when merging and the commit message
should be edited to succinctly summarize what the PR achieves instead of
enumerating every incremental change that went into the PR.
