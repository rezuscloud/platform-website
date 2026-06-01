# Multi-Cluster with Kamaji

RezusCloud uses [Kamaji](https://kamaji.clastix.io/) to run multiple Kubernetes control planes as pods within a single management cluster. Each tenant gets its own API server, controller manager, scheduler, and etcd, all running as containers on the management nodes. Worker nodes join from anywhere: edge bare metal, cloud VMs, or local machines.

## How it works

```
Management Cluster
┌─────────────────────────────────────────────────┐
│  Kamaji Controller                               │
│  ├── Tenant Control Plane: tenant-a              │
│  │   ├── kube-apiserver (pod)                    │
│  │   ├── kube-controller-manager (pod)           │
│  │   ├── kube-scheduler (pod)                    │
│  │   └── etcd (DataStore, possibly shared)       │
│  ├── Tenant Control Plane: tenant-b              │
│  │   └── ... (same layout)                       │
│  └── DataStore (etcd backing tenant planes)       │
│                                                   │
│  Worker Node Pool (joined to any tenant)           │
│  ├── Worker joining tenant-a                      │
│  └── Worker joining tenant-b                      │
└─────────────────────────────────────────────────┘
```

## Tenant lifecycle

1. Define a `RezusTenantConfig` CRD with the tenant specification
2. Kamaji provisions the control plane as pods
3. Worker nodes join the tenant's API server endpoint
4. Day 2 operations via `kubectl` with the tenant's kubeconfig

## Benefits

- **Isolation**: each tenant has its own API server and etcd
- **Density**: control planes share management cluster compute
- **Flexibility**: workers can be bare metal, cloud VMs, or edge devices
- **Simplicity**: one management cluster, many tenant clusters

## Source

This content is adapted from [rezusctl documentation](https://github.com/rezuscloud/rezusctl/blob/main/docs/multi-cluster.md) in the GitHub repository.

<!-- source: rezusctl:docs/multi-cluster.md -->
