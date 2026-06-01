# Multi-Cluster

RezusCloud can run multiple independent Kubernetes clusters from a single management cluster. Each tenant cluster has its own API server, scheduler, controller manager, and etcd, all running as pods inside the management cluster. Worker nodes join from anywhere: edge devices, cloud VMs, or bare metal servers.

This is useful when you need isolation between environments (production, staging, development) or between teams, without maintaining separate physical infrastructure for each.

## How it works

RezusCloud uses [Kamaji](https://kamaji.clastix.io/) to host tenant control planes as pods. The management cluster provides the compute for these control planes, while worker nodes connect to their assigned tenant's API server.

```mermaid
flowchart TB
    subgraph MC ["Management Cluster"]
        KAM[Kamaji Controller]
        subgraph TA ["Tenant: production"]
            API1[kube-apiserver]
            CM1[controller-manager]
            S1[scheduler]
        end
        subgraph TB2 ["Tenant: staging"]
            API2[kube-apiserver]
            CM2[controller-manager]
            S2[scheduler]
        end
        DS[(Shared etcd<br/>DataStore)]
    end

    subgraph W ["Worker Nodes"]
        WP[Production workers]
        WS[Staging workers]
    end

    KAM --> TA
    KAM --> TB2
    TA --> DS
    TB2 --> DS
    API1 --> WP
    API2 --> WS
```

## Tenant lifecycle

Setting up a new tenant is a single CRD application:

```bash
# Apply the tenant configuration
kubectl apply -f tenant-production.yaml
```

```yaml
apiVersion: rezuscloud.io/v1alpha1
kind: RezusTenantConfig
metadata:
  name: production
spec:
  controlPlane:
    version: "1.30"
    replicas: 1
  networking:
    podCIDR: "10.200.0.0/16"
    serviceCIDR: "10.100.0.0/16"
```

```mermaid
flowchart LR
    A[Define RezusTenantConfig] --> B[Kamaji provisions<br/>control plane pods]
    B --> C[Worker nodes join<br/>tenant API server]
    C --> D[kubectl with<br/>tenant kubeconfig]
```

## Benefits

- **Isolation**: each tenant has its own API server and etcd. A misconfiguration in one tenant cannot affect another
- **Density**: control planes share management cluster compute instead of requiring dedicated hardware
- **Flexibility**: workers can be anywhere: bare metal, cloud VMs, or edge devices
- **Simplicity**: one management cluster to monitor and maintain, many tenant clusters to use

<!-- source: rezusctl:docs/multi-cluster.md -->
