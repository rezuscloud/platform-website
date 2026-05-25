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

	if externalPath != "" {
		store, err = docs.NewStore(externalPath)
		if err != nil {
			log.Printf("docs: failed to load from %s: %v", externalPath, err)
		}
	}

	if store == nil {
		// Try embedded docs
		embedded := docs.EmbeddedDocs()
		store, err = docs.NewEmbeddedStore(embedded)
		if err != nil {
			log.Printf("docs: failed to load embedded docs: %v", err)
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

// DocsIndex renders the documentation landing page with repo listing.
func DocsIndex(c *fiber.Ctx) error {
	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	repos := DocsStore.Repos()
	return render(c, pages.DocsIndexPage(repos, DocsStore))
}

// DocsRepoIndex renders the doc index for a specific repo.
func DocsRepoIndex(c *fiber.Ctx) error {
	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	repoName := c.Params("repo")
	docsByCategory := DocsStore.DocsByCategory(repoName)
	if len(docsByCategory) == 0 {
		return c.Status(http.StatusNotFound).SendString("Repository not found")
	}

	// Find the repo config for display name
	var repo docs.RepoConfig
	for _, r := range docs.Registry {
		if r.Name == repoName {
			repo = r
			break
		}
	}

	return render(c, pages.DocsRepoPage(repo, docsByCategory, DocsStore))
}

// DocsPage renders a single documentation page.
func DocsPage(c *fiber.Ctx) error {
	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	repoName := c.Params("repo")
	// Accept path with or without .md extension
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

	// Get sibling docs for sidebar
	docsByCategory := DocsStore.DocsByCategory(repoName)

	var repo docs.RepoConfig
	for _, r := range docs.Registry {
		if r.Name == repoName {
			repo = r
			break
		}
	}

	return render(c, pages.DocsDetailPage(doc, repo, docsByCategory))
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
