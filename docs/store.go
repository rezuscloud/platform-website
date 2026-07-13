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
	// Path is the relative path from the docs root (e.g. "what-is-rezuscloud.md").
	Path string

	// Title is extracted from the first H1 in the markdown.
	Title string

	// HTML is the rendered markdown content.
	HTML string

	// Category is the top-level directory (e.g. "tutorials", "concepts").
	Category string

	// CategoryOrder controls sidebar display order. Lower = first.
	CategoryOrder int

	// GitHubURL links to the source file on GitHub (view).
	GitHubURL string

	// GitHubEditURL links to edit the file on GitHub.
	GitHubEditURL string
}

// categoryOrder defines the sidebar display order for categories.
// Follows the Diátaxis framework: Tutorials → How-to → Reference → Concepts.
var categoryOrder = map[string]int{
	"":           1,
	"tutorials":  2,
	"how-to":     3,
	"reference":  4,
	"concepts":   5,
	"operations": 6,
	"adr":        7,
}

// categoryDisplayNames maps directory names to sidebar headings.
var categoryDisplayNames = map[string]string{
	"":           "Overview",
	"tutorials":  "Tutorials",
	"how-to":     "How-to Guides",
	"reference":  "Reference",
	"concepts":   "Concepts",
	"operations": "Operations",
	"adr":        "Architecture Decisions",
	// Legacy mappings (for backward compatibility with older doc trees).
	"integrations": "Integrations",
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
type Store struct {
	mu sync.RWMutex

	// docs maps path → Doc (e.g. "what-is-rezuscloud.md" → Doc)
	docs map[string]Doc

	// orderedPaths is a sorted list of all doc paths.
	orderedPaths []string

	// basePath is the root directory containing doc files.
	basePath string
}

// NewStore creates a store that reads docs from basePath.
// basePath should contain category directories with markdown files:
//
//	basePath/
//	  what-is-rezuscloud.md
//	  concepts/
//	    architecture.md
//	  reference/
//	    cli.md
//	  adr/
//	    0001-foo.md
func NewStore(basePath string) (*Store, error) {
	s := &Store{
		docs:     make(map[string]Doc),
		basePath: basePath,
	}
	if err := s.loadFromDisk(); err != nil {
		return nil, err
	}
	return s, nil
}

// NewEmbeddedStore creates a store from an embedded filesystem.
func NewEmbeddedStore(fsys fs.FS) (*Store, error) {
	s := &Store{
		docs: make(map[string]Doc),
	}
	if err := s.loadFromFS(fsys); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) loadFromDisk() error {
	return filepath.WalkDir(s.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		relPath := strings.TrimPrefix(path, s.basePath+"/")
		if !shouldIndex(relPath) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		s.addDoc(relPath, data)
		return nil
	})
}

func (s *Store) loadFromFS(fsys fs.FS) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		if !shouldIndex(path) {
			return nil
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return nil
		}
		s.addDoc(path, data)
		return nil
	})
}

// shouldIndex reports whether a markdown file should be exposed in the docs UI.
// ADRs are kept in the repos as decision records but are not surfaced as
// end-user documentation.
//
// The versions/ subtree contains multi-version docs managed by fetch-docs.sh
// and read by the versioned Store (#110). Until the versioned Store is
// implemented, the legacy Store must not index these — they would collide
// with the backward-compat docs/external/ copies.
func shouldIndex(relPath string) bool {
	for _, seg := range strings.Split(relPath, "/") {
		if seg == "adr" {
			return false
		}
		if seg == "versions" {
			return false
		}
	}
	return true
}

// addDoc indexes a single markdown file.
//
// Path conventions:
//   - in-tree files (e.g. "what-is-rezuscloud.md") are served
//     at their natural path with no source attribution unless the file
//     carries an explicit <!-- source: repo:path --> comment.
//   - files under external/<repo>/... are fetched from the named repo by
//     scripts/fetch-docs.sh. The external/<repo>/ prefix is stripped from
//     the served URL, and the repo is recorded as the source for GitHub
//     view/edit links. The website is about rezuscloud; calling product
//     docs "external" in the URL would be misleading.
func (s *Store) addDoc(relPath string, data []byte) {
	title := ExtractTitle(data)
	if title == "" {
		base := strings.TrimSuffix(filepath.Base(relPath), ".md")
		title = strings.ReplaceAll(base, "-", " ")
		title = strings.ReplaceAll(title, "_", " ")
	}

	htmlContent, err := Render(data)
	if err != nil {
		return
	}

	// Default: source comes from an explicit comment in the file.
	repoName, sourcePath := parseSource(data)

	// Override: if the file lives under external/<repo>/..., the filesystem
	// path is authoritative. Strip the prefix from the served URL.
	if parts := strings.SplitN(relPath, "/", 3); len(parts) == 3 && parts[0] == "external" {
		repoName = parts[1]
		sourcePath = parts[2]
		relPath = parts[2]
	}

	githubURL := ""
	githubEditURL := ""
	if repoName != "" {
		repo := findRepo(repoName)
		if repo != nil {
			githubURL = repo.GitHubBaseURL() + "/" + sourcePath
			githubEditURL = repo.GitHubEditURL() + "/" + sourcePath
		}
	}

	category := ""
	if idx := strings.Index(relPath, "/"); idx >= 0 {
		category = relPath[:idx]
	}

	catOrder := 999
	if order, ok := categoryOrder[category]; ok {
		catOrder = order
	}

	s.docs[relPath] = Doc{
		Path:          relPath,
		Title:         title,
		HTML:          htmlContent,
		Category:      category,
		CategoryOrder: catOrder,
		GitHubURL:     githubURL,
		GitHubEditURL: githubEditURL,
	}

	s.rebuildIndex()
}

// parseSource extracts repo and path from a source comment in the markdown.
// Format: <!-- source: repo-name:path/within/repo/docs/file.md -->
func parseSource(data []byte) (repoName, sourcePath string) {
	text := string(data)
	idx := strings.LastIndex(text, "<!-- source:")
	if idx == -1 {
		return "", ""
	}
	line := text[idx:]
	line = strings.TrimPrefix(line, "<!-- source:")
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, "-->")
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return parts[0], ""
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
	var result []Doc
	for _, p := range s.orderedPaths {
		result = append(result, s.docs[p])
	}
	// Sort by category order, then path
	sort.Slice(result, func(i, j int) bool {
		di, dj := result[i], result[j]
		if di.CategoryOrder != dj.CategoryOrder {
			return di.CategoryOrder < dj.CategoryOrder
		}
		return di.Path < dj.Path
	})
	return result
}

// Categories returns categories in display order.
func (s *Store) Categories() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := make(map[string]bool)
	var result []string
	for _, p := range s.orderedPaths {
		d := s.docs[p]
		if !seen[d.Category] {
			seen[d.Category] = true
			result = append(result, d.Category)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		oi, oki := categoryOrder[result[i]]
		if !oki {
			oi = 999
		}
		oj, okj := categoryOrder[result[j]]
		if !okj {
			oj = 999
		}
		return oi < oj
	})
	return result
}

// DocsByCategory returns docs grouped by category, in display order.
func (s *Store) DocsByCategory() map[string][]Doc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string][]Doc)
	for _, p := range s.orderedPaths {
		d := s.docs[p]
		result[d.Category] = append(result[d.Category], d)
	}
	for cat := range result {
		sort.Slice(result[cat], func(i, j int) bool {
			return result[cat][i].Path < result[cat][j].Path
		})
	}
	return result
}

// DocCount returns the total number of docs.
func (s *Store) DocCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.docs)
}
