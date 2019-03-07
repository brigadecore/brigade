---
title: 'Tutorial 5: Writing Efficient Pipelines'
description: 'Advanced tutorial: Writing efficient pipelines'
section: intro
---

# Advanced tutorial: Writing efficient pipelines

_FIXME: This tutorial could use some love_

This advanced tutorial begins where [Tutorial 4][part4] left off. We’ll be parallelizing a few operations in brigade.js into separate jobs so the job can fail faster and run as fast as possible.

If you haven’t recently completed Tutorials 1–4, we strongly encourage you to review these so that your example project matches the one described below.

Here are a few things to help you:

- You can run as many jobs as you like in each event handler. We did just one, but you can do many.
- Feel free to use functions and objects in your scripts if you want to break things down into smaller parts.
- To get advanced flow control, take a look at `waitgroup`, which allows you to run a batch of jobs concurrently, then wait for them all to finish.
- DO NOT try to load external JavaScript modules (like Node modules). This is not supported.

If you've gone through this article, then you're probably scratching your head on what to [read next][readnext].

[readnext]: ../readnext