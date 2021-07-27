# Brigade Docs

* Browse the [docs on GitHub](https://github.com/brigadecore/brigade/tree/v2/docs/content)
* Browse the [v2 staging docs website](https://v2--brigade-docs.netlify.app)

---

## Development

The docs site is rendered using the [Hugo](https://gohugo.io/) static site generator, and a custom theme which borrows from the [Porter](https://github.com/getporter/porter) and [Helm](https://github.com/helm/helm-www) projects.

Commits to the v2 branch are auto deployed to the staging site via Netlify.

## Weights

Currently, the weights configured in the `config.toml` file specify ordering of each section in the main sidebar menu.

Weight values on individual documents are meant to specify ordering of pagination (previous and next navigation) within each doc's section.

This is fairly unwieldy, so if any docs/Hugo pros have knowledge on how to simplify, we'd love the contribution!
