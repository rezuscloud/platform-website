# CLI Reference

## Global Options

| Flag | Description |
|---|---|
| `--config` | Path to RezusCloudConfig file |
| `--kubeconfig` | Path to kubeconfig for management cluster |
| `--verbose` | Enable verbose logging |
| `--dry-run` | Print actions without executing |

## Commands

### `rezusctl boot`

Provision infrastructure and bootstrap the management cluster.

```bash
rezusctl boot --config rezuscloud-config.yaml
```

Idempotent: re-running applies diffs. Creates missing resources, updates changed ones, skips unchanged.

### `rezusctl join`

Generate worker node configuration for joining a tenant cluster.

```bash
rezusctl join --tenant my-tenant --output worker-config.yaml
```

### `rezusctl version`

Print the CLI version, git commit, and build time.

### `rezusctl status`

Show the current state of all managed resources.

## Environment Variables

| Variable | Description |
|---|---|
| `KUBECONFIG` | Default kubeconfig path |
| `REZUSCLOUD_CONFIG` | Default config file path |
| `OCI_CONFIG` | OCI SDK config file path |

## Source

This content is adapted from [rezusctl documentation](https://github.com/rezuscloud/rezusctl/blob/main/docs/cli-reference.md) in the GitHub repository.

<!-- source: rezusctl:docs/cli-reference.md -->
