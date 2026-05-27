package docs

import (
	"io/fs"
	"os"
	"path/filepath"
)

// GetDocFS returns the filesystem for reading docs.
// Priority:
//  1. DOCS_PATH env var (explicit path)
//  2. docs/external/ relative to working directory (build-time fetch)
//  3. nil (no docs available)
func GetDocFS() fs.FS {
	// Explicit override
	if path := os.Getenv("DOCS_PATH"); path != "" {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return os.DirFS(path)
		}
	}

	// Default: docs/external/ relative to working directory.
	// This is populated by scripts/fetch-docs.sh at build time
	// and available in both CI-built binaries and Docker containers.
	defaultPath := filepath.Join("docs", "external")
	if info, err := os.Stat(defaultPath); err == nil && info.IsDir() {
		return os.DirFS(defaultPath)
	}

	return nil
}
