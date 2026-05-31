package handlers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/docs"
	"github.com/rezuscloud/platform-website/views/pages"
)

// DocsStore is set during setup. Nil means docs are not available.
var DocsStore *docs.Store

// SetupDocs initializes the documentation store.
func SetupDocs(externalPath string) {
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
		repos := store.Repos()
		total := 0
		for _, r := range repos {
			count := store.RepoDocCount(r.Name)
			total += count
			log.Printf("docs: indexed %d pages from %s", count, r.DisplayName)
		}
		log.Printf("docs: %d pages across %d repos", total, len(repos))
	} else {
		log.Printf("docs: no documentation available")
	}

	DocsStore = store
}

// DocsIndex redirects to the first available doc page.
// Standard docs sites never show a project picker.
func DocsIndex(c *fiber.Ctx) error {
	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	allDocs := DocsStore.AllDocs()
	if len(allDocs) == 0 {
		return c.Status(http.StatusNotFound).SendString("No documentation available")
	}

	first := allDocs[0]
	return c.Redirect("/docs/"+first.RepoName+"/"+trimExt(first.Path), http.StatusMovedPermanently)
}

// DocsRepoIndex redirects to the first doc in a specific repo.
func DocsRepoIndex(c *fiber.Ctx) error {
	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	repoName := c.Params("repo")
	docsByCategory := DocsStore.DocsByCategory(repoName)
	if len(docsByCategory) == 0 {
		return c.Status(http.StatusNotFound).SendString("Repository not found")
	}

	// Find first doc
	for _, catDocs := range docsByCategory {
		if len(catDocs) > 0 {
			first := catDocs[0]
			return c.Redirect("/docs/"+first.RepoName+"/"+trimExt(first.Path), http.StatusMovedPermanently)
		}
	}

	return c.Status(http.StatusNotFound).SendString("No documents found")
}

// DocsPage renders a single documentation page with the standard layout:
// unified sidebar, content, right TOC, edit link, prev/next.
func DocsPage(c *fiber.Ctx) error {
	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	repoName := c.Params("repo")
	docPath := c.Params("*")
	if docPath == "" {
		return DocsRepoIndex(c)
	}
	if !hasSuffix(docPath, ".md") {
		docPath += ".md"
	}

	doc, found := DocsStore.Get(repoName, docPath)
	if !found {
		return c.Status(http.StatusNotFound).SendString("Document not found")
	}

	// Extract headings for right TOC
	headings := docs.ExtractHeadings(doc.HTML)

	// Build flat ordered list for prev/next
	allDocs := DocsStore.AllDocs()
	var prev, next *docs.Doc
	for i, d := range allDocs {
		if d.RepoName == doc.RepoName && d.Path == doc.Path {
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

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func trimExt(path string) string {
	if len(path) > 3 && path[len(path)-3:] == ".md" {
		return path[:len(path)-3]
	}
	return path
}
