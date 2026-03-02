# AGENTS.md

Guidelines for agentic coding agents operating in this repository.

## Project Overview

This is a marketing website for RezusCloud Enterprise Kubernetes Platform, built with:
- **Backend**: Go 1.24 + Fiber v2 web framework
- **Templating**: templ (type-safe HTML templates that generate Go code)
- **CSS**: Tailwind CSS v4 with class-based dark mode
- **Frontend**: HTMX 2.0.6 for progressive enhancement
- **Tests**: Playwright for visual regression and functional tests

## Build Commands

```bash
# Generate templ files (required after any .templ file changes)
templ generate

# Build CSS
npm run build:css
# or: npx @tailwindcss/cli -i input.css -o assets/styles.css --minify

# Build binary
CGO_ENABLED=0 go build -o bin/server .

# Full build (generate + css + build)
make build
```

## Development Commands

```bash
# Run development server
make dev
# or: go run .

# Live reload with all watchers (templ, server, tailwind)
make live

# Watch CSS only
npm run watch:css

# Watch templ only
templ generate --watch
```

## Lint & Format Commands

```bash
# Format templ files
templ fmt .

# Format Go code
go fmt ./...

# Check templ formatting (CI uses this)
templ fmt -fail .

# Check Go formatting
test -z "$(gofmt -l .)"

# Run Go vet
go vet ./...
```

## Test Commands

Tests are located in the `tests/` directory using Playwright.

```bash
# Run all tests
cd tests && npm test

# Run visual regression tests only
cd tests && npm run test:visual

# Run functional tests only
cd tests && npm run test:functional

# Run a single test file
cd tests && npx playwright test visual.spec.ts

# Run a single test by name
cd tests && npx playwright test -g "loads successfully"

# Run specific project
cd tests && npx playwright test --project=functional

# Update visual snapshots
cd tests && npm run test:update

# Open Playwright UI
cd tests && npm run test:ui

# View test report
cd tests && npm run report
```

**Note**: The server must be running on `http://localhost:3000` before running tests.

## Code Style Guidelines

### Go Code

- **Imports**: Group imports in this order:
  1. Standard library (e.g., `log`, `net`, `os`)
  2. External packages (e.g., `github.com/gofiber/fiber/v2`)
  3. Internal packages (e.g., `github.com/rezuscloud/platform-website/handlers`)
- **Error handling**: Use `log.Fatalf` for startup errors, return errors for handlers
- **Naming**: Use PascalCase for exported functions, camelCase for internal
- **Package names**: Single word, lowercase (e.g., `handlers`, `views`, `sections`)
- **Comments**: Add doc comments for exported functions

### templ Templates

- **File naming**: Use snake_case (e.g., `hero.templ`, `getstarted.templ`)
- **Generated files**: `*_templ.go` files are auto-generated, do not edit them
- **Component pattern**: Create reusable components as templ functions with parameters:
  ```templ
  templ featureCard(title string, description string, icon templ.Component) {
    // component content
  }
  ```
- **Children pattern**: Use `{ children... }` for wrapper components
- **Imports**: Import at the top of the file after package declaration

### Tailwind CSS

- **Dark mode**: Uses class strategy with `.dark` on `<html>` element
- **Custom variant**: Defined in `input.css` as `@custom-variant dark (&:where(.dark, .dark *));`
- **Color scheme**: Primary colors are cyan/blue gradients
- **Responsive**: Use `sm:`, `md:`, `lg:` prefixes for breakpoints
- **Do not edit**: `assets/styles.css` is generated, edit `input.css` instead

### TypeScript/Playwright Tests

- **Imports**: Use `import { test, expect, Page } from '@playwright/test';`
- **Constants**: Define at top of file (e.g., `BASE_URL`, `SECTIONS`)
- **Helper functions**: Mark as `async` and accept `Page` parameter
- **Test structure**: Group related tests with `test.describe()`
- **Snapshot naming**: `{section}-{viewport}-{theme}.png` format

## Project Structure

```
platform-website/
├── main.go              # Application entry point
├── handlers/
│   └── pages.go         # HTTP handlers (Home, Section)
├── views/
│   ├── layout.templ     # Base layout, Nav, Footer
│   ├── pages/
│   │   └── home.templ   # Page compositions
│   └── sections/        # Reusable section components
│       ├── hero.templ
│       ├── features.templ
│       └── ...
├── assets/
│   ├── js/htmx.min.js   # Vendored HTMX
│   └── styles.css       # Generated CSS (gitignored)
├── input.css            # Tailwind entry point
├── tests/               # Playwright tests
│   ├── playwright.config.ts
│   ├── visual.spec.ts
│   └── functional.spec.ts
└── Makefile             # Build automation
```

## Key Patterns

### Adding a New Section

1. Create `views/sections/newsection.templ`
2. Add `@templ Kids()` call in `views/pages/home.templ`
3. Add section to the `sectionMap` in `handlers/pages.go`
4. Add section ID to `SECTIONS` array in test files
5. Run `templ generate`

### HTTP Handlers

Handlers follow this pattern:
```go
func HandlerName(c *fiber.Ctx) error {
    return render(c, component)
}
```

### Theme Management

Theme is managed client-side via localStorage and the `dark` class on `<html>`. The `setTheme` helper in tests handles this.

## Pre-commit Checklist

1. `templ generate` - Regenerate templ files
2. `templ fmt .` - Format templ files
3. `go fmt ./...` - Format Go files
4. `go vet ./...` - Run Go vet
5. `npm run build:css` - Build CSS
6. Run tests if applicable

## Environment

- `PORT`: Server port (default: 3000)
