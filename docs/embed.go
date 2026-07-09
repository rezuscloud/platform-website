package docs

import (
	"io/fs"
	"os"
	"path/filepath"
)

// GetDocFS returns the filesystem for reading docs.
// Priority:
//  1. DOCS_PATH env var (explicit path)
//  2. docs/ relative to working directory (website-authored docs)
//  3. nil (no docs available)
func GetDocFS() fs.FS {
	// Explicit override
	if path := os.Getenv("DOCS_PATH"); path != "" {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return os.DirFS(path)
		}
	}

	// Default: docs/ directory containing website-authored documentation.
	// Organized by topic: tutorials/, concepts/, reference/, adr/
	defaultPath := filepath.Join("docs")
	if info, err := os.Stat(defaultPath); err == nil && info.IsDir() {
		return os.DirFS(defaultPath)
	}

	return nil
}
