package docs

import (
	"encoding/json"
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

// Manifest describes the available documentation versions.
// Written by scripts/fetch-docs.sh into docs/versions/manifest.json.
type Manifest struct {
	// Latest is the most recent release tag (e.g. "v0.0.1-106").
	Latest string `json:"latest"`

	// Next is the alias for unreleased main-branch docs ("next").
	Next string `json:"next"`

	// Versions lists all available versions, sorted newest-first.
	Versions []VersionEntry `json:"versions"`
}

// VersionEntry describes a single documentation version.
type VersionEntry struct {
	// Version is the tag name (e.g. "v0.0.1-106") or "next".
	Version string `json:"version"`

	// Minor is the major.minor group (e.g. "0.0").
	Minor string `json:"minor,omitempty"`

	// Date is the release date (ISO 8601).
	Date string `json:"date,omitempty"`

	// Label is a human-readable label (e.g. "Unreleased (main)").
	Label string `json:"label,omitempty"`

	// IsNext marks the unreleased main-branch entry.
	IsNext bool `json:"isNext,omitempty"`
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
//
// Two loading modes:
//
//   - Versioned mode: a docs/versions/manifest.json is present. Version-specific
//     docs are loaded from docs/versions/<tag>/, keyed per version. In-tree docs
//     (root *.md, contributing/) are shared across all versions.
//
//   - Legacy mode: no manifest. All docs (external/ + in-tree) are loaded into a
//     single flat map. This preserves backward compatibility with images built
//     before versioned docs.
type Store struct {
	mu sync.RWMutex

	// manifest describes available versions. nil in legacy mode.
	manifest *Manifest

	// versionDocs maps version tag → (stripped-path → Doc) for version-specific docs.
	// Only populated in versioned mode.
	versionDocs map[string]map[string]Doc

	// sharedDocs holds in-tree docs (stripped-path → Doc) that are the same across
	// all versions. In legacy mode, this holds ALL docs (external + in-tree).
	sharedDocs    map[string]Doc
	orderedShared []string

	basePath string
}

// NewStore creates a store that reads docs from basePath.
func NewStore(basePath string) (*Store, error) {
	s := &Store{
		sharedDocs:  make(map[string]Doc),
		versionDocs: make(map[string]map[string]Doc),
		basePath:    basePath,
	}
	if err := s.loadFromDisk(); err != nil {
		return nil, err
	}
	return s, nil
}

// NewEmbeddedStore creates a store from an embedded filesystem.
func NewEmbeddedStore(fsys fs.FS) (*Store, error) {
	s := &Store{
		sharedDocs:  make(map[string]Doc),
		versionDocs: make(map[string]map[string]Doc),
	}
	if err := s.loadFromFS(fsys); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) loadFromDisk() error {
	return s.loadFromFS(os.DirFS(s.basePath))
}

func (s *Store) loadFromFS(fsys fs.FS) error {
	// Try versioned mode: read manifest
	if manifest, err := s.loadManifest(fsys); err == nil && manifest != nil {
		s.manifest = manifest
		s.loadVersionedDocs(fsys)
		s.loadSharedDocs(fsys)
		return nil
	}

	// Legacy mode: no manifest — load all docs into sharedDocs
	s.loadLegacyDocs(fsys)
	return nil
}

// loadManifest reads versions/manifest.json from the filesystem.
func (s *Store) loadManifest(fsys fs.FS) (*Manifest, error) {
	data, err := fs.ReadFile(fsys, "versions/manifest.json")
	if err != nil {
		return nil, err
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if len(m.Versions) == 0 {
		return nil, fs.ErrNotExist
	}
	return &m, nil
}

// loadVersionedDocs walks docs/versions/<tag>/ for each version in the manifest.
func (s *Store) loadVersionedDocs(fsys fs.FS) {
	for _, v := range s.manifest.Versions {
		versionDir := "versions/" + v.Version
		docs := make(map[string]Doc)

		// GitHub ref: use the tag for releases, "main" for next
		ghRef := v.Version
		if v.IsNext {
			ghRef = "main"
		}

		fs.WalkDir(fsys, versionDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
				return nil
			}
			// Strip the versions/<tag>/ prefix BEFORE shouldIndex,
			// otherwise shouldIndex rejects the path (versions segment).
			relPath := strings.TrimPrefix(path, versionDir+"/")
			if !shouldIndex(relPath) {
				return nil
			}
			data, rerr := fs.ReadFile(fsys, path)
			if rerr != nil {
				return nil
			}
			doc := buildDoc(relPath, data, "rezuscloud", relPath, ghRef)
			docs[relPath] = doc
			return nil
		})

		s.versionDocs[v.Version] = docs
	}
}

// loadSharedDocs walks in-tree docs (root *.md, contributing/) that are shared
// across all versions. Excludes versions/, external/, adr/.
func (s *Store) loadSharedDocs(fsys fs.FS) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		// Skip versioned and external subtrees (handled separately)
		if isExcludedPath(path) {
			return nil
		}
		if !shouldIndex(path) {
			return nil
		}
		data, rerr := fs.ReadFile(fsys, path)
		if rerr != nil {
			return nil
		}
		s.addSharedDoc(path, data)
		return nil
	})
	s.rebuildSharedIndex()
}

// loadLegacyDocs loads all docs (external/ + in-tree) into sharedDocs.
// Used when no manifest is present (backward compatibility).
func (s *Store) loadLegacyDocs(fsys fs.FS) {
	fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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
		s.addSharedDoc(path, data)
		return nil
	})
	s.rebuildSharedIndex()
}

// isExcludedPath reports whether a path is in a subtree that should not be
// loaded as shared/in-tree docs (versions/ and external/ are handled by
// versioned loading or legacy external stripping).
func isExcludedPath(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if seg == "versions" || seg == "external" {
			return true
		}
	}
	return false
}

// shouldIndex reports whether a markdown file should be exposed in the docs UI.
// ADRs are kept in the repos as decision records but are not surfaced as
// end-user documentation. The versions/ subtree is handled by versioned loading.
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

// addSharedDoc indexes a markdown file into sharedDocs.
// Handles external/<repo>/ prefix stripping for legacy mode.
func (s *Store) addSharedDoc(relPath string, data []byte) {
	// External/<repo>/ prefix stripping (legacy mode)
	repoName, sourcePath := "", relPath
	if parts := strings.SplitN(relPath, "/", 3); len(parts) == 3 && parts[0] == "external" {
		repoName = parts[1]
		sourcePath = parts[2]
		relPath = parts[2]
	}

	// Explicit source comment overrides
	if r, p := parseSource(data); r != "" {
		repoName = r
		sourcePath = p
	}

	doc := buildDoc(relPath, data, repoName, sourcePath, "")
	s.sharedDocs[relPath] = doc
}

// buildDoc constructs a Doc from markdown data, extracting title, rendering
// HTML, and computing GitHub URLs.
func buildDoc(relPath string, data []byte, repoName, sourcePath, ghRef string) Doc {
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
			ref := ghRef
			if ref == "" {
				ref = repo.VersionTag
				if ref == "" {
					ref = "main"
				}
			}
			githubURL = "https://github.com/rezuscloud/" + repo.Name + "/blob/" + ref + "/" + repo.DocsPath + "/" + sourcePath
			githubEditURL = "https://github.com/rezuscloud/" + repo.Name + "/edit/" + ref + "/" + repo.DocsPath + "/" + sourcePath
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

	return Doc{
		Path:          relPath,
		Title:         title,
		HTML:          htmlContent,
		Category:      category,
		CategoryOrder: catOrder,
		GitHubURL:     githubURL,
		GitHubEditURL: githubEditURL,
	}
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

func (s *Store) rebuildSharedIndex() {
	paths := make([]string, 0, len(s.sharedDocs))
	for p := range s.sharedDocs {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	s.orderedShared = paths
}

// ---------------------------------------------------------------------------
// Version resolution
// ---------------------------------------------------------------------------

// HasVersions reports whether the store is in versioned mode.
func (s *Store) HasVersions() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.manifest != nil && len(s.manifest.Versions) > 0
}

// LatestVersion returns the latest release tag, or "" in legacy mode.
func (s *Store) LatestVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.manifest != nil {
		return s.manifest.Latest
	}
	return ""
}

// AvailableVersions returns all version entries from the manifest.
func (s *Store) AvailableVersions() []VersionEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.manifest == nil {
		return nil
	}
	return s.manifest.Versions
}

// ResolveVersion resolves a version alias to an actual version key.
// "latest" → manifest.Latest, "next" → manifest.Next, explicit → validated.
// Returns an error for unknown versions. In legacy mode, returns "".
func (s *Store) ResolveVersion(v string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.manifest == nil {
		return "", nil // legacy mode — no versioning
	}

	switch v {
	case "", "latest":
		return s.manifest.Latest, nil
	case "next":
		return s.manifest.Next, nil
	default:
		// Validate against manifest
		for _, ve := range s.manifest.Versions {
			if ve.Version == v {
				return v, nil
			}
		}
		return "", errUnknownVersion{v}
	}
}

// errUnknownVersion is returned when a version is not in the manifest.
type errUnknownVersion struct{ version string }

func (e errUnknownVersion) Error() string { return "unknown version: " + e.version }

// ---------------------------------------------------------------------------
// Document lookup
// ---------------------------------------------------------------------------

// Get is a legacy lookup that resolves to the latest version.
// Prefer GetVersioned for version-specific access.
func (s *Store) Get(path string) (Doc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	latest := ""
	if s.manifest != nil {
		latest = s.manifest.Latest
	}
	return s.getVersionedLocked(latest, path)
}

// GetVersioned looks up a doc by version and path. Falls back to shared/in-tree
// docs if the version-specific store doesn't have the path.
func (s *Store) GetVersioned(version, path string) (Doc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.getVersionedLocked(version, path)
}

func (s *Store) getVersionedLocked(version, path string) (Doc, bool) {
	// Try version-specific docs first
	if version != "" {
		if vd, ok := s.versionDocs[version]; ok {
			if d, found := vd[path]; found {
				return d, true
			}
		}
	}
	// Fallback: shared in-tree docs
	if d, found := s.sharedDocs[path]; found {
		return d, true
	}
	return Doc{}, false
}

// ---------------------------------------------------------------------------
// Listing (version-scoped)
// ---------------------------------------------------------------------------

// AllDocs returns all docs for a version: version-specific docs merged with
// shared in-tree docs, sorted by category order then path.
func (s *Store) AllDocs(version string) []Doc {
	s.mu.RLock()
	defer s.mu.RUnlock()

	combined := make(map[string]Doc)

	// Version-specific docs
	if version != "" {
		if vd, ok := s.versionDocs[version]; ok {
			for k, v := range vd {
				combined[k] = v
			}
		}
	}

	// Shared in-tree docs (version-specific takes precedence on collision)
	for k, v := range s.sharedDocs {
		if _, exists := combined[k]; !exists {
			combined[k] = v
		}
	}

	result := make([]Doc, 0, len(combined))
	for _, d := range combined {
		result = append(result, d)
	}
	sortDocs(result)
	return result
}

// Categories returns categories in display order for a version.
func (s *Store) Categories(version string) []string {
	docs := s.AllDocs(version)
	seen := make(map[string]bool)
	var result []string
	for _, d := range docs {
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

// DocsByCategory returns docs grouped by category for a version.
func (s *Store) DocsByCategory(version string) map[string][]Doc {
	docs := s.AllDocs(version)
	result := make(map[string][]Doc)
	for _, d := range docs {
		result[d.Category] = append(result[d.Category], d)
	}
	for cat := range result {
		sort.Slice(result[cat], func(i, j int) bool {
			return result[cat][i].Path < result[cat][j].Path
		})
	}
	return result
}

// sortDocs sorts docs by category order then path.
func sortDocs(docs []Doc) {
	sort.Slice(docs, func(i, j int) bool {
		if docs[i].CategoryOrder != docs[j].CategoryOrder {
			return docs[i].CategoryOrder < docs[j].CategoryOrder
		}
		return docs[i].Path < docs[j].Path
	})
}

// DocCount returns the total number of shared/in-tree docs.
func (s *Store) DocCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sharedDocs)
}
