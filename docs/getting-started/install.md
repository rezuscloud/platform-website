# Getting Started

## Prerequisites

| Requirement | Version | Check |
|---|---|---|
| Go | 1.26+ | `go version` |
| Kubernetes cluster | 1.30+ | — |
| Talos Linux | 1.12+ | — |
| Helm | 3.14+ | `helm version` |

For OCI cloud deployments:

| Requirement | Purpose |
|---|---|
| OCI account | Oracle Cloud Infrastructure |
| `~/.oci/config` | OCI SDK credentials |
| Cloudflare API token | DNS management |

## Install

Download the latest release from [GitHub](https://github.com/rezuscloud/rezusctl/releases):

```bash
# Linux (amd64)
curl -sL https://github.com/rezuscloud/rezusctl/releases/latest/download/rezusctl_linux_amd64.tar.gz | tar xz
sudo mv rezusctl /usr/local/bin/

# macOS (arm64)
curl -sL https://github.com/rezuscloud/rezusctl/releases/latest/download/rezusctl_darwin_arm64.tar.gz | tar xz
```

Verify the install:

```bash
rezusctl version
```

## Bootstrap a cluster

```bash
# Provision infrastructure and bootstrap the management cluster
rezusctl boot --config rezuscloud-config.yaml

# The boot command is idempotent: re-running applies diffs
# Creates missing resources, updates changed ones, skips unchanged
```

## Next steps

- [Architecture](/docs/concepts/architecture): understand how the platform fits together
- [Multi-Cluster](/docs/concepts/multi-cluster): run multiple Kubernetes control planes
- [CLI Reference](/docs/reference/cli): command-line options and flags

<!-- source: rezusctl:docs/getting-started.md -->
