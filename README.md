# RezusCloud Platform Website

The landing page for RezusCloud — your personal cloud. Built with Go Fiber, templ, HTMX, Alpine.js, and Tailwind CSS, running on the Dapr runtime.

## Overview

This is a single-page website that presents the **Personal Cloud** narrative: one person with any hardware and any ISP can run their own cloud infrastructure. The site draws the analogy from the personal computer revolution (mainframes → PCs) to today (cloud providers → personal cloud). The visual design evokes retro computing with CRT effects, monospace typography, and a cream/terminal color palette.

- **Single-page scrolling layout** with 11 content sections
- **Retro design system** — IBM Plex Mono + VT323 fonts, CRT scanlines, phosphor glow, beveled borders
- **Dark/light theme toggle** — cream background (light) / terminal green (dark), localStorage persistence
- **Alpine.js** for client-side interactivity (theme, mobile menu)
- **HTMX** for server-side interactivity (section updates)
- **Progressive enhancement** — works without JavaScript, enhanced with it
- **Dapr sidecar integration** for microservices building blocks
- **Multi-architecture container images** (amd64/arm64)

## Tech Stack

| Category | Technology | Version |
|----------|------------|---------|
| Runtime | Go | 1.24+ |
| Web Framework | Fiber v2 | 2.52.6 |
| Templating | templ | latest |
| CSS Framework | Tailwind CSS | v4 |
| Server Interactivity | HTMX | 2.0.6 |
| Client State | Alpine.js | 3.x |
| Fonts | IBM Plex Mono, VT323 | self-hosted woff2 |
| Container Runtime | Dapr | 1.15.3 |
| Base Image | distroless/static-debian12 | nonroot |

## Architecture

### Application Architecture

```mermaid
graph TB
    subgraph Pod["Pod: platform-website"]
        App["platform-website<br/>(Go Fiber app)<br/>Port: 3000<br/>Listens: tcp (IPv4 + IPv6)"]
        Dapr["daprd sidecar<br/>(Dapr runtime)<br/>Built-in components:<br/>- Kubernetes secrets<br/>- Service invocation"]
        App <--> Dapr
    end
```

### Request Flow

```mermaid
flowchart LR
    Browser["Browser"] --> Gateway["Gateway API (Cilium)"]
    Gateway --> HTTPRoute["HTTPRoute"]
    HTTPRoute --> Service["Service"]
    Service --> Pod["Pod"]
    Pod --> DaprSidecar["Dapr sidecar (optional)"]
    DaprSidecar --> DaprComponents["Dapr components (state, pub/sub)"]
```

### Project Structure

```
platform-website/
├── main.go                     # Application entry point
├── go.mod                      # Go module definition
├── go.sum                      # Go dependencies lock
├── input.css                   # Tailwind CSS entry point
├── package.json                # npm scripts for Tailwind CLI
├── Dockerfile                  # Multi-stage container build
├── Makefile                    # Build automation commands
├── README.md                   # This file
├── AGENTS.md                   # Guidelines for AI coding agents
├── .github/
│   └── workflows/
│       └── ci.yml              # GitHub Actions CI/CD
├── handlers/
│   ├── pages.go                # HTTP handlers (Home, Section)
│   └── api.go                  # API handlers (Version)
├── version/
│   └── version.go              # Version info (injected at build)
├── views/
│   ├── layout.templ            # Base HTML layout, Nav, Footer
│   ├── layout_templ.go         # Generated Go code
│   ├── components/             # Reusable components (placeholder)
│   ├── pages/
│   │   ├── home.templ          # Home page composition
│   │   └── home_templ.go       # Generated Go code
│   └── sections/
│       ├── hero.templ
│       ├── challenge.templ
│       ├── architecture.templ
│       ├── features.templ
│       ├── networking.templ
│       ├── edge.templ
│       ├── services.templ
│       ├── comparison.templ
│       ├── usecases.templ
│       ├── techstack.templ
│       └── getstarted.templ
├── assets/
│   ├── js/
│   │   ├── htmx.min.js         # Vendored HTMX 2.0.6
│   │   └── alpine.min.js       # Vendored Alpine.js 3.x
│   ├── img/
│   │   ├── icon.svg            # SVG favicon
│   │   ├── favicon.ico         # ICO favicon
│   │   ├── apple-touch-icon.png
│   │   ├── icon-192.png        # PWA icon
│   │   └── icon-512.png        # PWA icon
│   ├── manifest.webmanifest    # PWA manifest
│   └── styles.css              # Generated Tailwind CSS (gitignored)
└── tests/
    ├── integration_test.go     # Layer 2: goquery tests
    └── e2e_test.go             # Layer 3: chromedp tests
```

## Design Decisions

### 1. Go Fiber over net/http

Fiber provides excellent performance with its fasthttp foundation and offers a clean API for middleware composition. The middleware stack includes:

- `recover.New()` - Panic recovery
- `logger.New()` - Request logging
- `compress.New()` - Response compression

### 2. templ over html/template

templ offers type-safe HTML templating with Go code generation, eliminating runtime template errors and providing IDE support. Unlike `html/template`, templ:

- Generates Go code at build time
- Catches template errors during compilation
- Provides proper IDE autocompletion
- Enables component-based architecture

### 3. Tailwind CSS v4 with class strategy

The dark mode uses Tailwind's `class` strategy (not `media`) for explicit theme control:

```css
@custom-variant dark (&:where(.dark, .dark *));
```

Theme is managed by adding/removing the `dark` class on `<html>` and persisted to localStorage.

### 4. HTMX for Progressive Enhancement

The application supports both full-page renders and partial section updates:

| Route | Handler | Description |
|-------|---------|-------------|
| `GET /` | `handlers.Home` | Full page with all sections |
| `GET /sections/:name` | `handlers.Section` | Individual section for HTMX swaps |
| `GET /api/version` | `handlers.APIVersion` | JSON with version, gitCommit, buildTime |
| `GET /manifest.webmanifest` | Static file | PWA manifest |

This enables future enhancements like lazy-loading sections or animated transitions.

### 5. Dual-stack Network Listening

The server uses `net.Listen("tcp", ":3000")` instead of `tcp6` to listen on both IPv4 and IPv6:

```go
ln, err := net.Listen("tcp", addr)
```

This is required because Dapr sidecar communicates with the app via localhost (IPv4), while cluster traffic uses IPv6.

### 6. Dapr Runtime Integration

The application runs with Dapr sidecar injection for microservices capabilities:

**Annotations applied:**
```yaml
dapr.io/enabled: "true"
dapr.io/app-id: "platform-website"
dapr.io/app-port: "3000"
dapr.io/continue-on-failed-component-init: "true"
```

The `continue-on-failed-component-init` annotation allows Dapr to start even when the built-in Kubernetes secret store fails (occurs in distroless containers without kubeconfig).

### 7. Multi-stage Docker Build

The Dockerfile uses three stages for optimal image size:

1. **tailwind** (node:22-alpine) - Builds minified CSS
2. **builder** (golang:1.24-alpine) - Generates templ and compiles Go binary
3. **production** (distroless/static-debian12:nonroot) - Minimal runtime

Final image size: ~15MB

### 8. Distroless Base Image

Using `gcr.io/distroless/static-debian12:nonroot` provides:

- Minimal attack surface (no shell, package manager)
- Reduced CVE exposure
- Non-root execution by default
- Compatible with static Go binaries

## Content Sections

The website consists of 11 sections, each as a separate templ component:

| Section | ID | Purpose |
|---------|-----|---------|
| Hero | `#hero` | "Your Personal Cloud" manifesto + PC revolution analogy |
| Challenge | `#challenge` | "The Mainframe Moment" — then vs now parallel |
| Architecture | `#architecture` | "How It Works" — personal cloud blueprint |
| Features | `#features` | "What You Get" — personal benefits |
| Networking | `#networking` | "Always Connected" — any ISP, any network |
| Edge | `#edge` | "Runs on Anything" — old laptop, Raspberry Pi |
| Services | `#services` | "What You Can Run" — bundled software |
| Comparison | `#comparison` | "Own vs Rent" — personal decision |
| Use Cases | `#usecases` | "What Will You Build?" — inspiration |
| Tech Stack | `#techstack` | "What's Inside" — spec sheet |
| Get Started | `#getstarted` | "Start Your Cloud" — unboxing experience |

## Development

### Prerequisites

- Go 1.24+
- Node.js 22+
- templ CLI (`go install github.com/a-h/templ/cmd/templ@latest`)

### Local Development

```bash
# Install dependencies
npm install

# Download vendored JS libraries (HTMX, Alpine.js)
make vendor

# Generate templ files
templ generate

# Build CSS
npm run build:css

# Run the server
go run .
```

### Watch Mode

```bash
# Terminal 1: Watch CSS changes
npm run watch:css

# Terminal 2: Watch templ changes
templ generate --watch

# Terminal 3: Run server
go run .
```

### Building

```bash
# Generate all
templ generate
npm run build:css

# Build binary
CGO_ENABLED=0 go build -o bin/server .
```

### Docker Build

```bash
docker build -t platform-website .
docker run -p 3000:3000 platform-website
```

**Build Arguments:**

| Argument | Default | Description |
|----------|---------|-------------|
| `VERSION` | `dev` | Application version (e.g., `1.0.0`) |
| `GIT_COMMIT` | `unknown` | Git commit hash |
| `BUILD_TIME` | `unknown` | Build timestamp |

```bash
docker build \
  --build-arg VERSION=1.0.0 \
  --build-arg GIT_COMMIT=$(git rev-parse HEAD) \
  --build-arg BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ) \
  -t platform-website .
```

## CI/CD Pipeline

### GitHub Actions Workflows

The repository uses three workflows:

- `ci.yml` runs on pushes and pull requests to `master`
- `semantic-release.yml` runs after successful CI on `master` and creates the next semver tag
- `release.yml` runs on version tags (`v*`) and publishes release artifacts and container images

**CI (`.github/workflows/ci.yml`)**
1. Set up Go 1.24 and Node.js 22
2. Generate templ files and build CSS through the shared setup action
3. Check templ and Go formatting
4. Run `go vet`
5. Build the binary
6. Run handler tests and HTML integration tests on amd64 and arm64
7. Run chromedp E2E tests on amd64 against a real container image
8. For pull requests, publish a preview image tagged `pr-<number>-<sha>` to GHCR

**Release (`.github/workflows/release.yml`)**
1. Triggered by semver tags such as `v0.1.0`
2. Runs GoReleaser
3. Publishes multi-architecture images to GHCR
4. Publishes semver tags plus `latest`

### Container Registry

Images are published to:

```
ghcr.io/rezuscloud/platform-website:pr-<number>-<sha>
ghcr.io/rezuscloud/platform-website:v<version>
ghcr.io/rezuscloud/platform-website:v<major>
ghcr.io/rezuscloud/platform-website:v<major>.<minor>
ghcr.io/rezuscloud/platform-website:latest
```

## Kubernetes Deployment

### GitOps Source Of Truth

Application deployment is owned by `../k8s-config/apps/platform-website/`, not by this repository and not by a Terraform module. That directory contains:

- `namespace.yaml` - the `platform-website` namespace with `baseline` Pod Security labels
- `application.yaml` - the KubeVela `Application` that defines the website container, probes, Dapr annotations, and container security settings
- `httproute.yaml` - the Gateway API route for `rezus.cloud` and `www.rezus.cloud`
- `imagerepository.yaml` - Flux image scanning for `ghcr.io/rezuscloud/platform-website`
- `imagepolicy.yaml` - Flux semver selection for the newest allowed production tag
- `imageupdateautomation.yaml` - Flux automation that commits the selected tag back into `k8s-config`
- `preview/` - Flux preview-environment automation for open pull requests

The production image in `application.yaml` is intentionally pinned to a versioned tag such as `ghcr.io/rezuscloud/platform-website:v0.0.5` and annotated with a Flux image policy setter. Production does not deploy from `latest` directly.

### Production Rollout Flow

1. A change lands on `master` in `platform-website`
2. CI passes and `semantic-release.yml` may create a new semver tag when the merged commits warrant a release
3. `release.yml` publishes a multi-arch image to GHCR
4. Flux in `k8s-config` scans GHCR through `ImageRepository`
5. `ImagePolicy` selects the newest matching semver tag
6. `ImageUpdateAutomation` updates `k8s-config/apps/platform-website/application.yaml` with that tag and commits the change to `main`
7. Flux reconciles the updated manifests in the cluster
8. KubeVela materializes the `Deployment` and `Service`, and the separately-managed `HTTPRoute` keeps traffic pointed at the service

This means application code changes flow through this repository first, but the cluster-side deployment truth stays in `k8s-config`.

### Preview Environments

Pull requests also have GitOps-managed preview environments defined in `../k8s-config/apps/platform-website/preview/`.

- `rsip.yaml` watches GitHub pull requests for `rezuscloud/platform-website`
- `resourceset.yaml` creates one preview namespace per PR, named `platform-website-pr<id>`
- each preview `Application` uses image tag `ghcr.io/rezuscloud/platform-website:pr-<id>-<sha>`
- each preview `HTTPRoute` is exposed at `https://pr<id>.dev.rezus.cloud`
- `dev-wildcard-cert.yaml` provides the wildcard certificate for `*.dev.rezus.cloud`

Preview environments are generated reactively by Flux; they are not hand-authored per pull request.

### Request Path

```
Internet (IPv4) → 92.4.174.87:443 → OCI DNAT → 10.0.10.119:443
                                                ↓
Internet (IPv6) → [2603:c027:...]:443 ──────────┘
                                    Envoy (0.0.0.0:443, [::]:443)
                                           ↓
                                    HTTPRoute (rezus.cloud)
                                           ↓
                                     platform-website:3000 (ClusterIP)
                                            ↓
                                     Pod (Go Fiber + Dapr sidecar)
```

### Resources Created

| Resource | Type | Purpose |
|----------|------|---------|
| Namespace | v1 | Dedicated app namespace with `baseline` Pod Security labels |
| Application | core.oam.dev/v1beta1 | KubeVela definition for the website workload |
| Deployment | apps/v1 | Generated by KubeVela from the `Application` |
| Service | v1 | Generated by KubeVela on port 3000 |
| HTTPRoute | gateway.networking.k8s.io/v1 | Gateway API routing for production hostnames |

### Gateway API Configuration

The HTTPRoute routes `rezus.cloud` and `www.rezus.cloud` to the platform-website service:

```yaml
spec:
  hostnames:
    - rezus.cloud
    - www.rezus.cloud
  rules:
    - backendRefs:
        - name: platform-website
          port: 3000
```

Gateway API runs in **hostNetwork mode** — Cilium Envoy binds directly to host ports 80/443.
No LoadBalancer Service is created; traffic reaches Envoy on node IPs directly.
See `talos-iac/manifest/cni/README.md` for full CNI architecture.

### Pod Security

The namespace uses `baseline` Pod Security labels because the Dapr sidecar does not satisfy the stricter `restricted` profile. The application manifest also applies explicit pod and container security contexts, including non-root execution, `RuntimeDefault` seccomp, a read-only root filesystem, and dropped Linux capabilities.

### Dapr Configuration

The website workload is annotated for Dapr sidecar injection in `application.yaml`:

```yaml
dapr.io/enabled: "true"
dapr.io/app-id: platform-website
dapr.io/app-port: "3000"
dapr.io/continue-on-failed-component-init: "true"
dapr.io/scheduler-host-address: " "
```

The Dapr control plane itself is shared cluster infrastructure and is deployed separately from the website application:

- **Namespace:** `dapr-system`
- **Components:** Operator, Placement, Scheduler, Sentry, Sidecar Injector
- **HA Mode:** 3 replicas for scheduler

### Configuration Changes

Operational deployment changes should be made in `k8s-config/apps/platform-website/`, for example:

- changing replicas, probes, Dapr annotations, or security settings in `application.yaml`
- changing production hostnames or routing in `httproute.yaml`
- adjusting the preview deployment behavior under `preview/`

Cluster-wide prerequisites such as the gateway, cert-manager, Flux, or the Dapr control plane belong to `talos-iac` or `k8s-iac`, not this repository.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `3000` | Server listen port |

## Key Implementation Notes

### Retro Design System

The visual identity is inspired by personal computers of the 1980s (Olivetti, Apple Macintosh, Commodore 64):

- **Colors**: Cream (`#f5f0e8`), phosphor green (`#00ff41`), amber (`#ffb000`), retro blue (`#4a9eff`), terminal (`#0a0a0a`)
- **Fonts**: IBM Plex Mono (body) + VT323 (headings/accents), self-hosted as woff2
- **CRT effects**: Scanline overlay, text glow, pixel borders, beveled edges
- **Tailwind v4 theme**: Custom colors, fonts, and utilities defined in `input.css` via `@theme` blocks
- **Dark mode**: Class strategy with `.dark` on `<html>`, toggled via Alpine.js with localStorage persistence

**Important**: Tailwind v4 font variables use `--font-display` / `--font-retro` (not `--font-family-*`). See `input.css` for reference.

### Theme Persistence

Theme state is managed client-side:

1. On page load, check `localStorage.getItem('theme')`
2. If not set, detect `prefers-color-scheme` media query
3. Apply `dark` class to `<html>` element
4. Toggle updates localStorage and class

### Static Assets

Assets are served with `CacheDuration: -1` (no caching) to ensure fresh content during development. In production, the Gateway API can add cache headers.

### HTMX Integration

HTMX is vendored at `assets/js/htmx.min.js` to avoid external dependencies. Future enhancements could include:

- Lazy-loading sections on scroll
- Form submissions without page reload
- Real-time updates via SSE/WebSocket

### Version API

The `/api/version` endpoint returns build information:

```json
{
  "version": "1.0.0",
  "gitCommit": "abc123def456",
  "buildTime": "2024-01-15T10:30:00Z"
}
```

Version information is injected at build time via ldflags (see Docker Build section).

### Memory Footprint

The application has minimal memory overhead:

- Go binary: ~10MB
- Distroless base: ~2MB
- Total image: ~15MB
- Runtime memory: ~20-30MB

## License

MIT License - See LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes (run `templ generate` if modifying templates)
4. Submit a pull request

## Links

- **Live Site:** https://rezus.cloud
- **Source:** https://github.com/rezuscloud/platform-website
- **Container Registry:** https://github.com/rezuscloud/platform-website/pkgs/container/platform-website
- **Platform Documentation:** See `docs/PLATFORM.md` in the talos repository
