# Architecture

## Overview

`rezusctl` is the single binary that manages the full lifecycle of a RezusCloud Personal Cloud: provisioning cloud infrastructure, bootstrapping [Talos Linux](https://www.talos.dev/) nodes, installing platform components, managing multi-cluster fleets, and providing a WebUI for Day 2 operations.

**rezusctl builds clusters. kubectl manages them.**

## Design Principles

1. **No Terraform dependency.** All cloud provisioning via native Go SDK calls. Resource types are code-generated from Terraform provider schemas for type safety.
2. **CRD-based reconciliation.** Two CRDs, `RezusCloudConfig` and `RezusTenantConfig`, are the source of truth. `kubectl` and the WebUI read and write CRDs.
3. **Kubernetes-native Day 2.** No custom REST API. The WebUI is a static SPA that calls the Kubernetes API server directly via token-paste auth (same model as `kubectl`).
4. **Docker-first development.** Validate orchestration locally before touching cloud resources.
5. **Kamaji for hosted control planes.** [Kamaji](https://kamaji.clastix.io/) is installed as a dependency, not reimplemented. Each tenant's API server runs as pods in the management cluster.
6. **Idempotent boot.** Re-running `rezusctl boot` applies diffs: creates missing resources, updates changed ones, skips unchanged.

## Component Topology

```
Builder's Laptop                     Management Cluster
┌───────────────────┐               ┌──────────────────────────────────────────┐
│  rezusctl boot    │──────────────>│  Talos nodes (OCI VMs or bare metal)     │
│  (one-shot CLI)   │               │  ├── Cilium CNI                          │
│                   │               │  ├── cert-manager                        │
│  rezusctl join    │               │  ├── external-dns                        │
│  (worker config)  │               │  ├── Kamaji controller                   │
│                   │               │  ├── Kamaji DataStore (etcd)             │
│  kubectl          │<── kubecon ─> │  ├── Tenant Control Planes (pods)        │
│  (Day 2 ops)      │               │  │   ├── kube-apiserver                  │
```

## Runtime Stack

The management cluster runs these components after `rezusctl boot`:

| Component | Purpose |
|---|---|
| Cilium | CNI, network policy, Gateway API |
| cert-manager | TLS certificate lifecycle |
| external-dns | DNS records for platform services |
| Kamaji | Hosted control planes for tenant clusters |
| Dapr | Building block APIs for platform services |

## Source

This content is adapted from [rezusctl documentation](https://github.com/rezuscloud/rezusctl/blob/main/docs/architecture.md) in the GitHub repository.

<!-- source: rezusctl:docs/architecture.md -->
