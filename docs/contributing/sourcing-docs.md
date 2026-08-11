# How documentation is sourced into the website

The platform-website renders user-facing documentation from the **project
wikis**, fetched at build time. The repositories carry only ADRs and repo-level
context, which are never served here.

## Sources

| Source | What it holds | Fetched to |
|---|---|---|
| [`rezuscloud.wiki`](https://github.com/rezuscloud/rezuscloud/wiki) | Product docs: tutorials, how-to, reference, concepts | `docs/external/rezuscloud/` |
| [`platform-website.wiki`](https://github.com/rezuscloud/platform-website/wiki) | Website-authored pages (product intro) | `docs/external/platform-website/` |

`scripts/fetch-docs.sh` clones both wikis (public, no auth) into `docs/external/`.
It runs in CI (`ci.yml` build + docker-pr jobs) and as a goreleaser `before` hook,
so production images always carry fresh docs. Doc content therefore deploys the
next time a release image is built (push to `master`): the docs are baked in at
build time, not fetched at runtime, so a wiki edit on its own does not roll a
new image.

## Why wikis

A wiki is a single HEAD of user-facing docs. Because it contains no ADRs and no
archived decision history, internal content **cannot leak** onto the public docs
site. The store's Diátaxis category allowlist (`docs/store.go`) is the second
line of defense: only `tutorials/`, `how-to/`, `reference/`, `concepts/`, `operations/`,
and root-level pages (e.g. the product intro, documentation standards) are
served — everything else is ignored.

ADRs stay co-located with the code in each repo's `docs/adr/` (deeply
cross-referenced; co-versioned with the implementation).

## Diátaxis taxonomy

The sidebar follows the [Diátaxis](https://diataxis.fr/) framework:

| Directory | Sidebar label | Purpose |
|-----------|---------------|---------|
| (root) | Overview | High-level introductions |
| `tutorials/` | Tutorials | Learning-oriented: step-by-step, assumes nothing |
| `how-to/` | How-to Guides | Task-oriented: solve a specific problem |
| `reference/` | Reference | Information-oriented: precise, complete, structured |
| `concepts/` | Concepts | Understanding-oriented: deep, discursive |
| `operations/` | Operations | Production deployment + runbooks |

## Local development

Run the fetch once, then `make dev`:

```bash
bash scripts/fetch-docs.sh   # populates docs/external/
make dev
```

Or point `DOCS_PATH` at a local wiki checkout for live editing:

```bash
export DOCS_PATH=/path/to/rezuscloud.wiki
make dev
```
