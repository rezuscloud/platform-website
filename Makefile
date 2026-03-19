.PHONY: generate fmt build build-debug build-pgo build-pgo-train dev live clean css vendor htmx alpine

VERSION ?= dev
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.gitCommit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME)
PGO_FILE ?= default.pgo

# Generate templ files to Go code
generate:
	templ generate

# Format templ files
fmt:
	templ fmt .
	go fmt ./...

# Build CSS
css:
	npx @tailwindcss/cli -i input.css -o assets/styles.css --minify

# Build the application (optimized)
build: generate css
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o bin/server .

# Build the application (debug - no optimizations)
build-debug: generate css
	CGO_ENABLED=0 go build -gcflags="all=-N -l" -o bin/server .

# Run the server
dev: generate css
	go run .

# Live reload (run all watchers in parallel)
live/templ:
	templ generate --watch --proxy="http://localhost:3000" --open-browser=false

live/server:
	go run github.com/air-verse/air@latest \
		--build.cmd "go build -o tmp/bin/main ." --build.bin "tmp/bin/main" --build.delay "100" \
		--build.exclude_dir "node_modules" \
		--build.include_ext "go" \
		--build.stop_on_error "false" \
		--misc.clean_on_exit true

live/tailwind:
	npx @tailwindcss/cli -i input.css -o assets/styles.css --watch

live:
	make -j3 live/templ live/server live/tailwind

# Download HTMX
htmx:
	curl -sL https://unpkg.com/htmx.org@2.0.6/dist/htmx.min.js -o assets/js/htmx.min.js

# Download Alpine.js
alpine:
	curl -sL https://cdn.jsdelivr.net/npm/alpinejs@3.x.x/dist/cdn.min.js -o assets/js/alpine.min.js

# Download all vendored JS libraries
vendor: htmx alpine

# Clean build artifacts
clean:
	rm -rf bin/ tmp/ assets/styles.css

# Build with PGO profile (requires default.pgo to exist)
build-pgo: generate css
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -pgo=$(PGO_FILE) -o bin/server .

# Build for profile collection (used to generate PGO profile)
build-pgo-train: generate css
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -trimpath -o bin/server .

# Collect PGO profile (run server with PPROF_ENABLED=true, then generate load)
# Example: make collect-pgo BENCHMARK_CMD="curl -s http://localhost:3000/ > /dev/null"
collect-pgo:
	@echo "Starting server for PGO profile collection..."
	@echo "Run your benchmark/workload, then Ctrl+C to stop."
	@echo "Profile will be saved to $(PGO_FILE)"
	CGO_ENABLED=0 go build -o bin/server-pgo-train .
	PPROF_ENABLED=true ./bin/server-pgo-train &
	@sleep 2
	@echo "Server started. Run your workload now..."
	@if [ -n "$(BENCHMARK_CMD)" ]; then \
		echo "Running benchmark: $(BENCHMARK_CMD)"; \
		for i in $$(seq 1 100); do $(BENCHMARK_CMD); done; \
	fi
	@echo "Collecting CPU profile..."
	go tool pprof -proto http://localhost:6060/debug/pprof/profile?seconds=30 > $(PGO_FILE)
	@pkill -f "bin/server-pgo-train" || true
	@echo "PGO profile saved to $(PGO_FILE)"
