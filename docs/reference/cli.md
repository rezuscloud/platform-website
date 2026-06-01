# CLI Reference

Complete reference for the `rezusctl` command-line tool.

## Global flags

| Flag | Description |
|---|---|
| `--config` | Path to RezusCloudConfig file |
| `--kubeconfig` | Path to kubeconfig for management cluster |
| `--verbose` | Enable verbose logging |
| `--dry-run` | Print actions without executing |

## Commands

### `rezusctl boot`

Provision infrastructure and bootstrap the management cluster. This is the primary command you run to create or update your Personal Cloud.

```bash
rezusctl boot --config rezuscloud-config.yaml
```

The command is idempotent. Running it again after a change to your config applies only the diff: creates new resources, updates changed ones, and leaves everything else untouched.

### `rezusctl join`

Generate a worker node configuration for joining a tenant cluster.

```bash
rezusctl join --tenant production --output worker-config.yaml
```

Apply the generated config on the target node to join it to the tenant's API server.

### `rezusctl version`

Print the CLI version, git commit hash, and build timestamp.

### `rezusctl status`

Display the current state of all managed resources: nodes, tenant clusters, platform components, and their health.

## Environment variables

| Variable | Description |
|---|---|
| `KUBECONFIG` | Default kubeconfig path for the management cluster |
| `REZUSCLOUD_CONFIG` | Default path to the RezusCloudConfig file |
| `OCI_CONFIG` | Path to OCI SDK configuration file |

<!-- source: rezusctl:docs/cli-reference.md -->
