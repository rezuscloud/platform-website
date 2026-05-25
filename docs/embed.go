package docs

import (
	"embed"
	"io/fs"
)

//go:embed all:external
var embeddedDocs embed.FS

// EmbeddedDocs returns the embedded filesystem rooted at the external/ directory.
func EmbeddedDocs() fs.FS {
	sub, _ := fs.Sub(embeddedDocs, "external")
	return sub
}
