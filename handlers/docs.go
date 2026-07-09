package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/docs"
	"github.com/rezuscloud/platform-website/views/pages"
)

// DocsStore is set during setup. Nil means docs are not available.
var DocsStore *docs.Store

// SetupDocs initializes the documentation store.
func SetupDocs() {
	var store *docs.Store
	var err error

	fs := docs.GetDocFS()
	if fs != nil {
		store, err = docs.NewEmbeddedStore(fs)
		if err != nil {
			log.Printf("docs: failed to load docs: %v", err)
		}
	}

	if store != nil {
		allDocs := store.AllDocs()
		log.Printf("docs: indexed %d pages across %d categories", len(allDocs), len(store.Categories()))
	} else {
		log.Printf("docs: no documentation available")
	}

	DocsStore = store
}

// DocsIndex redirects to the first available doc page.
func DocsIndex(c *fiber.Ctx) error {
	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	allDocs := DocsStore.AllDocs()
	if len(allDocs) == 0 {
		return c.Status(http.StatusNotFound).SendString("No documentation available")
	}

	return c.Redirect("/docs/"+trimExt(allDocs[0].Path), http.StatusMovedPermanently)
}

// DocsPage renders a single documentation page.
// Accepts paths like /docs/getting-started/install or /docs/concepts/architecture.
func DocsPage(c *fiber.Ctx) error {
	// Check for a redirect (renamed/moved docs) before anything else —
	// redirects don't depend on the docs store being loaded.
	if docPath := c.Params("*"); docPath != "" {
		if newPath := docs.Redirect(docPath); newPath != "" {
			return c.Redirect("/docs/"+newPath, http.StatusMovedPermanently)
		}
	}

	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	// Build the path from the wildcard parameter
	docPath := c.Params("*")
	if docPath == "" {
		return DocsIndex(c)
	}
	if !strings.HasSuffix(docPath, ".md") {
		docPath += ".md"
	}

	doc, found := DocsStore.Get(docPath)
	if !found {
		return c.Status(http.StatusNotFound).SendString("Document not found")
	}

	// Extract headings for right TOC
	headings := docs.ExtractHeadings(doc.HTML)

	// Build flat ordered list for prev/next
	allDocs := DocsStore.AllDocs()
	var prev, next *docs.Doc
	for i, d := range allDocs {
		if d.Path == doc.Path {
			if i > 0 {
				prev = &allDocs[i-1]
			}
			if i < len(allDocs)-1 {
				next = &allDocs[i+1]
			}
			break
		}
	}

	return render(c, pages.DocsDetailPage(doc, headings, prev, next, DocsStore))
}

func trimExt(path string) string {
	if len(path) > 3 && path[len(path)-3:] == ".md" {
		return path[:len(path)-3]
	}
	return path
}
