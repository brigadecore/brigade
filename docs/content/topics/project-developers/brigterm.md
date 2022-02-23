---
title: Brigterm
description: Using Brigade's text-based visualization utility
section: project-developers
weight: 5
aliases:
  - /brigterm
  - /topics/brigterm.md
  - /topics/project-developers/brigterm.md
---

# Brig term

Brigade offers a text-based user interface (TUI) for visualizing the activity
in a Brigade system. This can be a convenient utility for viewing all projects
and monitoring their events, as well as digging down into worker and job
logs for an event.

# How to use

To start, log in to the Brigade server that you wish to explore:

```console
$ brig login --server https://my-brigade-server.example.com
```

Then, simply invoke the following:

```console
$ brig term
```

The visualizer will expand to encompass the entire terminal window and you
should be greeted with a project listing on the main page.

# Sample views

Here is a sample view of the main project overview page:

![Project Overview](/img/brigterm_project_overview.png)

From this view, you can navigate to a given project to see its full details and
event listing using the arrow keys. Press the `Enter` key to select the project
you wish to view. Here we have selected the `first-job` project:

![Project view](/img/brigterm_first-job_project.png)

To view details for a specific event, select the ID you wish to explore. Here
we select the one event available to us:

![Event view](/img/brigterm_first-job_event.png)

Worker logs for the event can be seen by typing the `L` key:

![Worker logs](/img/brigterm_first-job_worker_logs.png)

You can also select any job to view its full details. Here we navigate to the
details for the `my-first-job` job:

![Job view](/img/brigterm_first-job_job.png)

Finally, job logs can be seen by typing the `L` key:

![Job logs](/img/brigterm_first-job_job_logs.png)

We hope this visualizer proves useful for tracking activity in a Brigade
instance. We're excited to hear about your experiences with it!
