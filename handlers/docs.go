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

// SetupDocs initializes the documentation store from the on-disk docs tree
// (fetched wikis + in-tree pages).
func SetupDocs() {
	var store *docs.Store
	fs := docs.GetDocFS()
	if fs != nil {
		s, err := docs.NewEmbeddedStore(fs)
		if err != nil {
			log.Printf("docs: failed to load docs: %v", err)
		} else {
			store = s
		}
	}

	if store != nil {
		log.Printf("docs: indexed %d pages across %d categories", store.DocCount(), len(store.Categories()))
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
// Accepts paths like /docs/tutorials/install-and-first-cluster.
// Renamed/moved docs (DocRedirects) are redirected to their new path.
func DocsPage(c *fiber.Ctx) error {
	docPath := c.Params("*")
	if docPath == "" {
		return DocsIndex(c)
	}

	// Redirect renamed/moved docs (301) before lookup.
	if newPath := docs.Redirect(docPath); newPath != "" {
		return c.Redirect("/docs/"+newPath, http.StatusMovedPermanently)
	}

	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	lookupPath := docPath
	if !strings.HasSuffix(lookupPath, ".md") {
		lookupPath += ".md"
	}

	doc, found := DocsStore.Get(lookupPath)
	if !found {
		// Category roots (/docs/how-to) have no index page — the breadcrumb
		// links them, so redirect to the section's first doc in sidebar order.
		// 302, not 301: the mapping is derived ("first doc of category") and
		// shifts as pages are added, so it must not be cached as permanent.
		if first, ok := DocsStore.FirstDocInCategory(strings.TrimSuffix(docPath, "/")); ok {
			return c.Redirect("/docs/"+trimExt(first), http.StatusFound)
		}
		return c.Status(http.StatusNotFound).SendString("Document not found")
	}

	headings := docs.ExtractHeadings(doc.HTML)

	// Build flat ordered list for prev/next.
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
	return strings.TrimSuffix(path, ".md")
}
