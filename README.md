# RezusCloud Platform Website

RezusCloud treats the homepage as a same-origin shell that composes three live application surfaces:

- `terminal-app`
- `mac-app`
- `linux-app`

The shell remains marketing-first, but the proof path is built around real Dapr primitives: service invocation, pub/sub, shared state, Redis-backed locking, and namespace-local deployment topology.

## Architecture

### Runtime modes

The repository builds one image. Runtime behavior is selected with `PLATFORM_MODE`:

- unset: local gateway for development and tests
- `shell`: homepage shell service
- `terminal`: terminal service
- `mac`: Mac inspection service
- `linux`: Linux execution service

The local gateway mounts the same routes under one process for fast verification:

- `/`
- `/shell/summary`
- `/apps/terminal`
- `/apps/terminal/embed`
- `/apps/mac`
- `/apps/mac/embed`
- `/apps/linux`
- `/apps/linux/embed`

### Dapr contract

The first vertical slice proves the architecture with one session-scoped flow:

1. User runs a command in `terminal-app`
2. `terminal-app` invokes `linux-app`
3. `linux-app` writes topology and artifact state
4. `linux-app` publishes an event to `homepage.events`
5. `mac-app` renders the artifact from shared state
6. `shell-app` observes the event and updates the proof rail

In cluster, that flow is backed by concrete namespace-local Dapr components:

- PostgreSQL V2 for `statestore`
- NATS JetStream for `pubsub`
- Redis for `configstore`
- Redis for `lockstore`

The shared state key is namespaced per session:

- `homepage/sessions/<id>/state`

### Cluster deployment topology

GitOps manifests live in `../k8s-config/apps/platform-website/`.

In-cluster, the deployment is split into two layers:

- `../k8s-config/platform/dapr/` installs the shared Dapr control plane in `dapr-system`
- `../k8s-config/apps/platform-website/` installs this application's namespace-local Dapr resources and backends

Production runs four KubeVela components from the same image behind one hostname:

- `platform-website-shell`
- `platform-website-terminal`
- `platform-website-mac`
- `platform-website-linux`

Each deployable has its own Dapr app ID:

- `platform-website-shell`
- `platform-website-terminal`
- `platform-website-mac`
- `platform-website-linux`

The production namespace `platform-website` also runs dedicated Dapr backends for this application:

- PostgreSQL for state
- NATS JetStream for pub/sub
- Redis for configuration and locking

The app code uses those concrete component names explicitly:

- `DAPR_STATE_STORE_NAME=statestore`
- `DAPR_PUBSUB_NAME=pubsub`
- `DAPR_LOCK_STORE_NAME=lockstore`

Preview environments use the same topology. Flux creates a fresh namespace `platform-website-pr<N>` with its own Dapr components and its own PostgreSQL, JetStream, and Redis backends for every pull request that carries the `preview-ready` label.

### Observability

Dapr telemetry is collected by a dedicated SigNoz `k8s-infra` chart (separate from the main SigNoz deployment):

- **Traces** — Dapr sidecars export OTLP gRPC traces directly to the SigNoz collector at `signoz-otel-collector.signoz.svc.cluster.local:4317`
- **Metrics** — a Deployment-mode OTEL collector scrapes sidecar metrics from port 9090 using label-based discovery
- **Logs** — a DaemonSet-mode OTEL collector tails Dapr sidecar logs from `/var/log/pods/`

All signals flow through the main SigNoz collector into ClickHouse. The k8s-infra chart configuration lives in `../k8s-config/infrastructure/signoz/`.

## Repo layout

```text
platform-website/
├── cmd/
│   ├── shell/
│   ├── terminal/
│   ├── mac/
│   └── linux/
├── internal/
│   ├── apps/
│   ├── platform/
│   └── server/
├── views/
│   ├── apps/
│   ├── pages/
│   └── layout.templ
├── tests/
├── assets/
├── input.css
├── Dockerfile
└── main.go
```

## Development

### Prerequisites

- Go 1.24+
- Node.js 22+
- `templ` CLI

### Local run

```bash
npm ci
templ generate
npm run build:css
go run .
```

Then open `http://localhost:3000`.

### Build split services locally

```bash
go build ./cmd/shell
go build ./cmd/terminal
go build ./cmd/mac
go build ./cmd/linux
```

### Tests

```bash
go test ./...
go test -tags=e2e ./tests/...
```
