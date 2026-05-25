package docs

import (
	"io/fs"
	"os"
)

// GetDocFS returns the filesystem for reading docs.
// Priority: DOCS_PATH env var → embedded docs (if any) → nil
func GetDocFS() fs.FS {
	// Check for external docs path (set in CI/Docker builds)
	if path := os.Getenv("DOCS_PATH"); path != "" {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return os.DirFS(path)
		}
	}

	// No docs available
	return nil
}
