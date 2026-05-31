package docs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Doc represents a single rendered documentation page.
type Doc struct {
	// RepoName is the GitHub repo this doc belongs to.
	RepoName string

	// RepoDisplayName is the human-readable repo name.
	RepoDisplayName string

	// Path is the relative path within the repo's docs directory.
	// e.g. "adr/0001-personal-cloud-identity.md"
	Path string

	// Title is extracted from the first H1 in the markdown.
	Title string

	// HTML is the rendered markdown content.
	HTML string

	// GitHubURL links to the source file on GitHub.
	GitHubURL string

	// Category groups docs (e.g. "adr", "" for root docs).
	Category string
}

// Store reads and caches documentation from the filesystem.
type Store struct {
	mu sync.RWMutex

	// docs maps "repoName/path" → Doc
	docs map[string]Doc

	// repoIndex maps repoName → list of Doc paths
	repoIndex map[string][]string

	// basePath is the root directory containing repo doc folders.
	basePath string
}

// NewStore creates a store that reads docs from basePath.
// basePath should contain one subdirectory per repo, each with
// the repo's docs/ contents.
//
//	layout:
//	  basePath/
//	    rezusctl/
//	      architecture.md
//	      versioning.md
//	      adr/
//	        0001-foo.md
//	    platform-website/
//	      adr/
//	        0001-bar.md
func NewStore(basePath string) (*Store, error) {
	s := &Store{
		docs:      make(map[string]Doc),
		repoIndex: make(map[string][]string),
		basePath:  basePath,
	}

	if err := s.loadAll(); err != nil {
		return nil, err
	}
	return s, nil
}

// NewEmbeddedStore creates a store from the embedded filesystem.
func NewEmbeddedStore(fsys fs.FS) (*Store, error) {
	s := &Store{
		docs:      make(map[string]Doc),
		repoIndex: make(map[string][]string),
	}

	for _, repo := range Registry {
		if err := s.loadRepoFromFS(fsys, repo); err != nil {
			continue // skip repos without docs
		}
	}

	return s, nil
}

func (s *Store) loadAll() error {
	for _, repo := range Registry {
		repoDir := filepath.Join(s.basePath, repo.Name)
		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			continue
		}
		if err := s.loadRepoFromDisk(repo, repoDir); err != nil {
			return fmt.Errorf("loading docs for %s: %w", repo.Name, err)
		}
	}
	return nil
}

func (s *Store) loadRepoFromDisk(repo RepoConfig, repoDir string) error {
	return filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		relPath := strings.TrimPrefix(path, repoDir+"/")
		s.addDoc(repo, relPath, data)
		return nil
	})
}

func (s *Store) loadRepoFromFS(fsys fs.FS, repo RepoConfig) error {
	dir := repo.Name
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return err
	}
	return s.walkFS(fsys, repo, dir, entries)
}

func (s *Store) walkFS(fsys fs.FS, repo RepoConfig, dir string, entries []fs.DirEntry) error {
	for _, entry := range entries {
		fullPath := dir + "/" + entry.Name()
		if entry.IsDir() {
			sub, err := fs.ReadDir(fsys, fullPath)
			if err != nil {
				continue
			}
			if err := s.walkFS(fsys, repo, fullPath, sub); err != nil {
				return err
			}
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		data, err := fs.ReadFile(fsys, fullPath)
		if err != nil {
			continue
		}
		relPath := strings.TrimPrefix(fullPath, repo.Name+"/")
		s.addDoc(repo, relPath, data)
	}
	return nil
}

func (s *Store) addDoc(repo RepoConfig, relPath string, data []byte) {
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

	category := ""
	if idx := strings.Index(relPath, "/"); idx >= 0 {
		category = relPath[:idx]
	}

	doc := Doc{
		RepoName:        repo.Name,
		RepoDisplayName: repo.DisplayName,
		Path:            relPath,
		Title:           title,
		HTML:            htmlContent,
		GitHubURL:       repo.GitHubBaseURL() + "/" + relPath,
		Category:        category,
	}

	key := repo.Name + "/" + relPath
	s.docs[key] = doc
	s.repoIndex[repo.Name] = append(s.repoIndex[repo.Name], relPath)

	if paths, ok := s.repoIndex[repo.Name]; ok {
		sort.Strings(paths)
		s.repoIndex[repo.Name] = paths
	}
}

// Get returns a doc by repo name and path.
func (s *Store) Get(repoName, path string) (Doc, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.docs[repoName+"/"+path]
	return d, ok
}

// ListByRepo returns all docs for a repo, sorted by path.
func (s *Store) ListByRepo(repoName string) []Doc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	paths, ok := s.repoIndex[repoName]
	if !ok {
		return nil
	}
	var result []Doc
	for _, p := range paths {
		if d, ok := s.docs[repoName+"/"+p]; ok {
			result = append(result, d)
		}
	}
	return result
}

// Repos returns all repos that have documentation.
func (s *Store) Repos() []RepoConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []RepoConfig
	for _, repo := range Registry {
		if _, ok := s.repoIndex[repo.Name]; ok {
			result = append(result, repo)
		}
	}
	return result
}

// CategoriesForRepo returns unique categories for a repo, in order of first appearance.
func (s *Store) CategoriesForRepo(repoName string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := make(map[string]bool)
	var result []string
	for _, p := range s.repoIndex[repoName] {
		cat := ""
		if idx := strings.Index(p, "/"); idx >= 0 {
			cat = p[:idx]
		}
		if !seen[cat] {
			seen[cat] = true
			result = append(result, cat)
		}
	}
	return result
}

// AllDocs returns all docs across all repos, ordered by repo name then path.
func (s *Store) AllDocs() []Doc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Doc
	for _, repo := range Registry {
		paths, ok := s.repoIndex[repo.Name]
		if !ok {
			continue
		}
		for _, p := range paths {
			if d, ok := s.docs[repo.Name+"/"+p]; ok {
				result = append(result, d)
			}
		}
	}
	return result
}

// DocsByCategory returns docs grouped by category for a repo.
func (s *Store) DocsByCategory(repoName string) map[string][]Doc {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string][]Doc)
	for _, p := range s.repoIndex[repoName] {
		d := s.docs[repoName+"/"+p]
		result[d.Category] = append(result[d.Category], d)
	}
	for cat := range result {
		sort.Slice(result[cat], func(i, j int) bool {
			return result[cat][i].Path < result[cat][j].Path
		})
	}
	return result
}

// RepoDocCount returns the number of docs for a repo.
func (s *Store) RepoDocCount(repoName string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.repoIndex[repoName])
}
