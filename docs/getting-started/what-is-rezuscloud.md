# What is RezusCloud?

RezusCloud is the Personal Cloud: a fully open source, free platform that delivers enterprise-grade infrastructure to anyone who needs it. Multi-site edge computing, distributed workloads, GPU scheduling, AI guardrailing and orchestration, access controls, audit trails, and more, all included.

## The Personal Cloud

A cloud you control instead of rent as a managed service. The same way personal computers made mainframe power accessible to individuals, the Personal Cloud makes cloud infrastructure accessible to anyone with hardware.

**Personal** means self-managed (vs. managed by a cloud provider), not individual. An enterprise with 200 engineers can run a Personal Cloud. Their Machine Room might be a rack in a data center instead of a Raspberry Pi on a shelf, but it is still theirs.

## The Machine Room

The physical infrastructure you control and operate. A rack in an office, a cluster of Raspberry Pis at home, a free OCI VM, a Hetzner dedicated server, or a closet full of gear. You control the hardware, the data, and the computation. No landlord, no lease, no eviction.

## The Golden Path

The tested, fully featured default configuration that covers 90% of real-world use cases. The platform takes away Kubernetes complexity by providing this path. Builders deploy on top of K8s without operating it.

## Why own your cloud

| Problem | Cloud rental | Personal Cloud |
|---|---|---|
| Traffic costs | Per-byte egress tolls | Data moves through encrypted tunnels between owned nodes |
| Denial of wallet | Per-request billing, attackers inflate your bill | No per-request billing. No bill to inflate |
| Egress fees | Provider charges for accessing your own files | Data stays on owned disks |
| Pricing changes | Vendor changes costs at any time | Hardware depreciation, electricity, or a flat lease are predictable for years |
| Idle resources | 30% of cloud spend is pure waste | Owned hardware at any utilization is free headroom |
| Vendor lock-in | Proprietary APIs, migration pain | Every component is open source and replaceable |

## Start with a Pi, scale to a rack

Same platform, same git push, same operational simplicity. A Raspberry Pi at home, a free OCI VM, a Hetzner dedicated server, or a rack in a closet. No traffic costs, no egress fees, no API burst surprises, no pricing policy changes, no idle resource waste, no lock-in.

<!-- source: platform-website:docs/adr/0001-personal-cloud-identity.md -->
