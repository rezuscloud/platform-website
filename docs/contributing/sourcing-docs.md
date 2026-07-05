# How to source RezusCloud docs into the website

The platform-website renders documentation from two sources:

1. **Website-authored pages** — `docs/` in this repo (marketing, product identity, website-specific ADRs).
2. **Product technical docs** — `docs/` in the [rezuscloud repo](https://github.com/rezuscloud/rezuscloud) (tutorials, how-to, reference, concepts, operations, ADRs).

## Diátaxis taxonomy

The sidebar follows the [Diátaxis framework](https://diataxis.fr/):

| Directory | Sidebar Label | Purpose |
|-----------|---------------|---------|
| `tutorials/` | Tutorials | Learning-oriented: step-by-step, assumes nothing |
| `how-to/` | How-to Guides | Task-oriented: solve a specific problem |
| `reference/` | Reference | Information-oriented: precise, complete, structured |
| `concepts/` | Concepts | Understanding-oriented: deep, discursive |
| `operations/` | Operations | Production deployment and runbooks |
| `adr/` | Architecture Decisions | Decision records |

## Local development

To render the full rezuscloud docs locally, set `DOCS_PATH` to point at the
rezuscloud docs directory:

```bash
# From the platform-website directory
export DOCS_PATH=/path/to/rezuscloud/docs
make dev
```

Or create a symlink so both doc trees are visible:

```bash
ln -s /path/to/rezuscloud/docs/tutorials docs/tutorials
ln -s /path/to/rezuscloud/docs/how-to docs/how-to
ln -s /path/to/rezuscloud/docs/reference docs/reference
ln -s /path/to/rezuscloud/docs/concepts docs/concepts
ln -s /path/to/rezuscloud/docs/operations docs/operations
ln -s /path/to/rezuscloud/docs/adr docs/adr
```

## Production deployment

In production, set `DOCS_PATH` to a directory containing the merged doc tree,
or use a build step that copies the rezuscloud docs into the container image.

## Source attribution

Every doc page should include a source comment at the end:

```html
<!-- source: rezuscloud:docs/concepts/components.md -->
```

This drives the "Edit on GitHub" link. The registry in `docs/registry.go` maps
repo names to GitHub URLs.
