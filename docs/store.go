package docs

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Doc represents a single rendered documentation page.
type Doc struct {
	// Path is the relative path from the docs root (e.g. "tutorials/install-and-first-cluster.md").
	Path string

	// Title is extracted from the first H1 in the markdown.
	Title string

	// HTML is the rendered markdown content.
	HTML string

	// Category is the top-level directory (e.g. "tutorials", "concepts", or ""
	// for root-level Overview pages).
	Category string

	// CategoryOrder controls sidebar display order. Lower = first.
	CategoryOrder int

	// GitHubURL links to the source (view). For wiki-sourced docs, the wiki.
	GitHubURL string

	// GitHubEditURL links to edit the source. For wiki-sourced docs, the wiki.
	GitHubEditURL string
}

// allowedCategories is the explicit Diátaxis allowlist. Only these top-level
// categories are surfaced as documentation; every other path (ADRs, wiki meta,
// repository READMEs, future additions) is ignored. This is authoritative — a
// fetched wiki that gains a new top-level directory never appears on the public
// docs site unless it is added here.
//
// "" is Overview (root-level pages such as the product introduction).
var allowedCategories = map[string]bool{
	"":           true,
	"tutorials":  true,
	"how-to":     true,
	"reference":  true,
	"concepts":   true,
	"operations": true,
}

// categoryOrder defines the sidebar display order, following Diátaxis:
// Overview → Tutorials → How-to → Reference → Concepts.
var categoryOrder = map[string]int{
	"":           1,
	"tutorials":  2,
	"how-to":     3,
	"reference":  4,
	"concepts":   5,
	"operations": 6,
}

// categoryDisplayNames maps categories to sidebar headings.
var categoryDisplayNames = map[string]string{
	"":           "Overview",
	"tutorials":  "Tutorials",
	"how-to":     "How-to Guides",
	"reference":  "Reference",
	"concepts":   "Concepts",
	"operations": "Operations",
}

// CategoryDisplayName returns the display name for a category directory.
func CategoryDisplayName(cat string) string {
	if name, ok := categoryDisplayNames[cat]; ok {
		return name
	}
	return strings.Title(cat)
}

// DocRedirects maps old doc paths (without .md) to new paths (without .md).
// Used when docs are renamed or moved during reorganizations so old URLs
// redirect (301) instead of 404.
var DocRedirects = map[string]string{
	// Diátaxis reorg (PR #169) — getting-started/ → tutorials/ + concepts/
	"getting-started/index":              "tutorials/install-and-first-cluster",
	"getting-started/multi-cluster":      "concepts/multi-cluster",
	"getting-started/what-is-rezuscloud": "what-is-rezuscloud",
	"getting-started/install":            "tutorials/install-and-first-cluster",
	// integrations/ → how-to/
	"integrations/home-assistant": "how-to/integrate-home-assistant",
}

// Redirect returns the new path for a renamed doc, or "" if no redirect exists.
func Redirect(oldPath string) string {
	return DocRedirects[oldPath]
}

// Store reads and caches documentation from the filesystem.
//
// Documentation is authored in the project wikis (rezuscloud.wiki,
// platform-website.wiki) and fetched at build time into docs/external/<repo>/ by
// scripts/fetch-docs.sh. In-tree pages (root *.md) are also indexed. Only the
// Diátaxis categories in allowedCategories are served.
type Store struct {
	mu           sync.RWMutex
	docs         map[string]Doc
	orderedPaths []string
}

// NewStore creates a store that reads docs from basePath.
func NewStore(basePath string) (*Store, error) {
	s := &Store{docs: make(map[string]Doc)}
	if err := s.loadFromFS(os.DirFS(basePath)); err != nil {
		return nil, err
	}
	return s, nil
}

// NewEmbeddedStore creates a store from an embedded/filesystem.
func NewEmbeddedStore(fsys fs.FS) (*Store, error) {
	s := &Store{docs: make(map[string]Doc)}
	if err := s.loadFromFS(fsys); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) loadFromFS(fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if !shouldIndex(path) {
			return nil
		}
		data, rerr := fs.ReadFile(fsys, path)
		if rerr != nil {
			return nil
		}
		s.addDoc(path, data)
		return nil
	})
}

// shouldIndex is a fast pre-filter that drops wiki/repository meta pages before
// they are read. The authoritative gate is the Diátaxis category allowlist
// applied in addDoc (after the external/<repo>/ prefix is stripped).
func shouldIndex(relPath string) bool {
	switch filepath.Base(relPath) {
	case "Home.md", "_Sidebar.md", "_Footer.md", "README.md":
		return false
	}
	return true
}

// addDoc indexes a markdown file. Files under external/<repo>/... are fetched
// from that repo's wiki; the prefix is stripped from the served path and the
// repo is recorded as the source for GitHub view/edit links.
func (s *Store) addDoc(relPath string, data []byte) {
	repoName, sourcePath := "", relPath
	if parts := strings.SplitN(relPath, "/", 3); len(parts) == 3 && parts[0] == "external" {
		repoName = parts[1]
		sourcePath = parts[2]
		relPath = parts[2]
	}

	category := ""
	if idx := strings.Index(relPath, "/"); idx >= 0 {
		category = relPath[:idx]
	}

	// Authoritative gate: only the Diátaxis categories are served.
	if !allowedCategories[category] {
		return
	}

	s.docs[relPath] = buildDoc(relPath, data, repoName, sourcePath)
	s.rebuildIndex()
}

// buildDoc constructs a Doc from markdown data, extracting the title, rendering
// HTML, and computing GitHub URLs.
func buildDoc(relPath string, data []byte, repoName, sourcePath string) Doc {
	title := ExtractTitle(data)
	if title == "" {
		base := strings.TrimSuffix(filepath.Base(relPath), ".md")
		title = strings.ReplaceAll(base, "-", " ")
		title = strings.ReplaceAll(title, "_", " ")
	}

	htmlContent, err := Render(data)
	if err != nil {
		htmlContent = "<p>Failed to render document.</p>"
	}

	githubURL := ""
	githubEditURL := ""
	if repoName != "" {
		repo := findRepo(repoName)
		if repo != nil {
			if repo.IsWiki {
				// Wikis are edited via their web UI; link to the wiki, not a
				// per-file blob URL (wiki page URLs don't preserve subdirs).
				githubURL = repo.GitHubBaseURL()
				githubEditURL = repo.GitHubEditURL()
			} else {
				githubURL = repo.GitHubBaseURL() + "/" + sourcePath
				githubEditURL = repo.GitHubEditURL() + "/" + sourcePath
			}
		}
	}

	category := ""
	if idx := strings.Index(relPath, "/"); idx >= 0 {
		category = relPath[:idx]
	}

	return Doc{
		Path:          relPath,
		Title:         title,
		HTML:          htmlContent,
		Category:      category,
		CategoryOrder: categoryOrder[category],
		GitHubURL:     githubURL,
		GitHubEditURL: githubEditURL,
	}
}

// findRepo finds a RepoConfig by name.
func findRepo(name string) *RepoConfig {
	for i := range Registry {
		if Registry[i].Name == name {
			return &Registry[i]
		}
	}
	return nil
}

func (s *Store) rebuildIndex() {
	paths := make([]string, 0, len(s.docs))
	for p := range s.docs {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	s.orderedPaths = paths
}

// Get returns a doc by its relative path.
func (s *Store) Get(path string) (Doc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.docs[path]
	return d, ok
}

// AllDocs returns all docs, sorted by category order then path.
func (s *Store) AllDocs() []Doc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Doc, 0, len(s.orderedPaths))
	for _, p := range s.orderedPaths {
		result = append(result, s.docs[p])
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].CategoryOrder != result[j].CategoryOrder {
			return result[i].CategoryOrder < result[j].CategoryOrder
		}
		return result[i].Path < result[j].Path
	})
	return result
}

// Categories returns categories in display order.
func (s *Store) Categories() []string {
	docs := s.AllDocs()
	seen := make(map[string]bool)
	var result []string
	for _, d := range docs {
		if !seen[d.Category] {
			seen[d.Category] = true
			result = append(result, d.Category)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return categoryOrder[result[i]] < categoryOrder[result[j]]
	})
	return result
}

// DocsByCategory returns docs grouped by category, in display order.
func (s *Store) DocsByCategory() map[string][]Doc {
	docs := s.AllDocs()
	result := make(map[string][]Doc)
	for _, d := range docs {
		result[d.Category] = append(result[d.Category], d)
	}
	return result
}

// DocCount returns the total number of docs.
func (s *Store) DocCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.docs)
}
