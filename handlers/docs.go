package handlers

import (
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/rezuscloud/platform-website/docs"
	"github.com/rezuscloud/platform-website/views/pages"
)

// DocsStore is set during setup. Nil means docs are not available.
var DocsStore *docs.Store

// versionTagPattern matches explicit version prefixes like "v0.0.1-106/", "v0.1.0/".
var versionTagPattern = regexp.MustCompile(`^v\d+\.\d+\.`)

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
		if store.HasVersions() {
			latest := store.LatestVersion()
			allDocs := store.AllDocs(latest)
			versions := store.AvailableVersions()
			log.Printf("docs: indexed %d pages across %d categories (versioned: %d versions, latest=%s)",
				len(allDocs), len(store.Categories(latest)), len(versions), latest)
		} else {
			allDocs := store.AllDocs("")
			log.Printf("docs: indexed %d pages across %d categories (legacy mode)",
				len(allDocs), len(store.Categories("")))
		}
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

	latest := DocsStore.LatestVersion()
	allDocs := DocsStore.AllDocs(latest)
	if len(allDocs) == 0 {
		return c.Status(http.StatusNotFound).SendString("No documentation available")
	}

	return c.Redirect("/docs/"+trimExt(allDocs[0].Path), http.StatusMovedPermanently)
}

// parseVersionPrefix splits a raw doc path into (versionAlias, remainingPath).
// Returns ("", fullPath) if no version prefix is present.
//
// Examples:
//
//	"v0.0.1-106/tutorials/install" → ("v0.0.1-106", "tutorials/install")
//	"latest/tutorials/install"      → ("latest", "tutorials/install")
//	"next/tutorials/install"        → ("next", "tutorials/install")
//	"tutorials/install"             → ("", "tutorials/install")
func parseVersionPrefix(rawPath string) (version, rest string) {
	idx := strings.Index(rawPath, "/")

	// No slash: check if entire path is a bare version alias
	if idx < 0 {
		if rawPath == "latest" || rawPath == "next" || versionTagPattern.MatchString(rawPath+"/") {
			return rawPath, ""
		}
		return "", rawPath
	}

	// Has slash: check first segment
	firstSeg := rawPath[:idx]
	if firstSeg == "latest" || firstSeg == "next" || versionTagPattern.MatchString(firstSeg+"/") {
		return firstSeg, rawPath[idx+1:]
	}
	return "", rawPath
}

// versionPrefixForURL returns the URL prefix for sidebar/breadcrumb links.
// Returns "" for unversioned pages so links stay at /docs/<path>.
func versionPrefixForURL(requestedVersion string) string {
	if requestedVersion == "" {
		return ""
	}
	return requestedVersion + "/"
}

// DocsPage renders a single documentation page.
// Accepts paths like /docs/tutorials/install, /docs/latest/concepts/architecture,
// /docs/v0.0.1-106/reference/cli.
func DocsPage(c *fiber.Ctx) error {
	rawPath := c.Params("*")
	if rawPath == "" {
		return DocsIndex(c)
	}

	// Parse version prefix (e.g. "latest/", "v0.0.1-106/", "next/")
	requestedVersion, docPath := parseVersionPrefix(rawPath)

	// Redirect check (version-aware) — redirects don't depend on the store.
	if newPath := docs.Redirect(docPath); newPath != "" {
		return c.Redirect("/docs/"+versionPrefixForURL(requestedVersion)+newPath, http.StatusMovedPermanently)
	}

	if DocsStore == nil {
		return c.Status(http.StatusNotFound).SendString("Documentation not available")
	}

	// Resolve version alias to actual version key
	resolvedVersion, err := DocsStore.ResolveVersion(requestedVersion)
	if err != nil {
		return versionNotFound(c, requestedVersion)
	}

	if docPath == "" {
		// Version root: redirect to first doc in this version
		allDocs := DocsStore.AllDocs(resolvedVersion)
		if len(allDocs) > 0 {
			return c.Redirect(
				"/docs/"+versionPrefixForURL(requestedVersion)+trimExt(allDocs[0].Path),
				http.StatusMovedPermanently,
			)
		}
		return c.Status(http.StatusNotFound).SendString("No documentation available")
	}

	// Look up the document
	lookupPath := docPath
	if !strings.HasSuffix(lookupPath, ".md") {
		lookupPath += ".md"
	}

	doc, found := DocsStore.GetVersioned(resolvedVersion, lookupPath)
	if !found {
		return docNotFound(c, requestedVersion, docPath)
	}

	// Build headings and prev/next for the resolved version
	headings := docs.ExtractHeadings(doc.HTML)
	allDocs := DocsStore.AllDocs(resolvedVersion)

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

	versionPrefix := versionPrefixForURL(requestedVersion)
	return render(c, pages.DocsDetailPage(doc, headings, prev, next, DocsStore, resolvedVersion, versionPrefix))
}

// versionNotFound renders a 404 for an unrecognized version.
func versionNotFound(c *fiber.Ctx, version string) error {
	c.Status(http.StatusNotFound)
	return c.SendString("Version " + version + " is not available. " +
		"<a href=\"/docs\">View latest documentation</a>.")
}

// docNotFound renders a 404 for a doc that doesn't exist in the current version.
func docNotFound(c *fiber.Ctx, requestedVersion, docPath string) error {
	// Check if the doc exists in the latest version — if so, suggest it.
	latest := DocsStore.LatestVersion()
	lookupPath := docPath
	if !strings.HasSuffix(lookupPath, ".md") {
		lookupPath += ".md"
	}

	c.Status(http.StatusNotFound)
	if _, exists := DocsStore.GetVersioned(latest, lookupPath); exists {
		return c.SendString("This page is not available in this version. " +
			"<a href=\"/docs/latest/" + docPath + "\">View in latest version</a>.")
	}
	return c.SendString("Document not found.")
}

func trimExt(path string) string {
	if len(path) > 3 && path[len(path)-3:] == ".md" {
		return path[:len(path)-3]
	}
	return path
}
