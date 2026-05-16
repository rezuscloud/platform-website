# RezusCloud Platform

The domain language for the RezusCloud platform, its website, and its positioning. This context covers product strategy, user-facing terminology, and the boundaries between concepts.

## Language

**Personal Cloud**:
A cloud you control instead of rent as a managed service. Enterprise-grade infrastructure running on personal devices, free cloud VMs, or dedicated servers. Not a scale descriptor, a paradigm shift: the same way personal computers made mainframe power accessible to individuals, the Personal Cloud makes cloud infrastructure accessible to anyone with hardware. Ownership of the platform, data, and computation as an extension of the self-hosted manifesto.
_Avoid_: private cloud (implies isolation, not ownership), home cloud (implies scale limit), self-hosted cloud (describes method, not paradigm)

**Machine Room**:
The physical infrastructure a user controls and operates. A rack in an office, a cluster of Pis at home, a free cloud VM, a dedicated server in a data center. The modern equivalent of the personal computer on your desk. You control the hardware, the data, and the computation. No landlord, no lease, no eviction.
_Avoid_: server room (implies dedicated facility), data center (implies enterprise facility), homelab (implies hobby)

**Golden Path**:
The tested, fully featured default configuration that covers 90% of real-world use cases. The platform takes away Kubernetes complexity by providing this path. Builders deploy on top of K8s without operating it.
_Avoid_: default config (too technical), best practices (too generic), recommended setup (too cautious)

**Builder**:
Anyone who uses the platform to deploy and operate infrastructure. A solo developer, an ML engineer, an ops lead, a platform team. The person who pushes to git.
_Avoid_: user (too generic), developer (excludes ops), customer (implies transaction)

## Relationships

- A **Personal Cloud** is composed of one or more **Machine Rooms** connected by encrypted tunnels
- The **Golden Path** is what makes a **Personal Cloud** operable without Kubernetes expertise
- A **Builder** operates a **Personal Cloud**, regardless of whether they are solo or part of an organization
- The **Machine Room** provides building blocks (compute, storage, networking, orchestration) that enable any containerized workload: AI inference, data pipelines, CI/CD, observability, edge computing

## Cloud pain points addressed

The Personal Cloud eliminates six structural problems with renting cloud infrastructure as a managed service:

1. **Traffic costs**: Data moves between owned nodes through encrypted tunnels. No per-byte egress tolls.
2. **Denial of wallet attacks**: No per-request billing. Attackers cannot inflate a bill that doesn't exist.
3. **Egress fees**: Data stays on owned disks. No provider charges for accessing your own files.
4. **Pricing policy changes**: No vendor can change costs on you. Hardware depreciation, electricity, or a flat server lease are predictable years ahead.
5. **Idle resource waste**: 30% of cloud spend is pure waste. Owned hardware at any utilization is free headroom, not money burned.
6. **Vendor lock-in**: Every component is open source and replaceable. Standard container protocols, standard orchestration.

## Example dialogue

> **Dev:** "Can an enterprise with 200 engineers use the **Personal Cloud**?"
> **Domain expert:** "Yes. 'Personal' means they own it, not that only one person uses it. Their **Machine Room** might be a rack in a data center instead of a Pi on a shelf, but it's still theirs."

## Flagged ambiguities

- "Personal" was initially read as "single-user." Resolved: personal means self-managed (vs. managed by a cloud provider), not individual. The parallel is the Personal Computer: you controlled the machine instead of renting mainframe time.
- "Enterprise" was initially excluded from the audience. Resolved: enterprises are a target. The anti-reference is enterprise *marketing* (procurement committees, sales calls), not enterprise *users*.
