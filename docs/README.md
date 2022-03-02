# docs.brigade.sh

[![Netlify Status](https://api.netlify.com/api/v1/badges/82538e29-5fcd-4196-8fa5-8de57cc096ed/deploy-status)](https://app.netlify.com/sites/brigade-docs/deploys)

> ⚠️&nbsp;&nbsp;If you're looking to browse the docs, you might prefer exploring the [docs.brigade.sh](https://docs.brigade.sh).
>
> 📚&nbsp;&nbsp;If you're looking to update the docs, you've come to the right place!
>
> 🏠&nbsp;&nbsp;For edits to the main [brigade.sh](https://brigade.sh) site, go [here](https://github.com/brigadecore/brigade-www).
>
> 📢&nbsp;&nbsp;For contributions to [blog.brigade.sh](https://blog.brigade.sh) go [here](https://github.com/brigadecore/blog).

## Development

> ⚠️&nbsp;&nbsp;If you're only looking to edit content, you can skip this
> section.

The site is a simple [Hugo](https://gohugo.io/) site built with some tweaks to
the [Techdoc](https://github.com/thingsym/hugo-theme-techdoc) theme.

> ⚠️&nbsp;&nbsp;The contents of `themes/hugo-theme-techdoc` are a git submodule.
> __Do not tweak the theme by modifying the contents of
> `themes/hugo-theme-techdoc`.__ Instead, copy files you wish to tweak to a
> parallel directory structure under the root of the repository and modify them
> there. Hugo will automatically favor such files when building the site.

After initially cloning this repository, be sure to run the following commands:

```shell
$ git submodule init
$ git submodule update
```

To run the website locally, you'll need to first
[install Hugo](https://gohugo.io/getting-started/installing/).

You can then run a local development server. This process will watch for changes
and automatically refresh content, styles, etc.:

```shell
$ cd docs
$ hugo serve
```

## Deployment

Changes are automatically deployed to
[Netlify](https://app.netlify.com/sites/brigade-docs) when merged to
the `main` branch.

Build logs can be found
[here](https://app.netlify.com/sites/brigade-docs/deploys).

## How to Add or Edit Content

Content changes are created via pull requests.

If you're creating a _new_ page, follow these steps:

1. Add a new markdown file to the `content/` directory. The directory structure
   within the `content/` directory describes the layout of the doc tree, so be
   sure to place your content in the right subdirectory. You may create a new
   subdirectory if applicable.

1. Add [front matter](https://gohugo.io/content-management/front-matter/) to the
   file using this format:

   ```yaml
   ---
   linkTitle: Optional title used in the doc tree, but NOT used as the page's heading
   title: The page's title/heading
   description: An optional byline
   section: Should match the name of the subdirectory the page is in
   weight: Numerical weight within the section; lower weights appear higher in the doc tree
   ---
   ```

   Example:

   ```yaml
   ---
   linkTitle: Quickstart
   title: A Brigade Quickstart
   section: intro
   weight: 3
   ---

1. Add the content below the `---` as markdown. The title MUST NOT be
   re-included as a header in the content.

1. Open a PR.
