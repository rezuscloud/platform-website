package handlers

import (
	"io"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/rezuscloud/platform-website/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// versionedTestFS creates a docs fixture with two versions + next + shared in-tree docs.
func versionedTestFS() fstest.MapFS {
	return fstest.MapFS{
		"versions/manifest.json": {
			Data: []byte(`{
  "latest": "v1.0.0",
  "next": "next",
  "versions": [
    {"version": "v1.0.0", "minor": "1.0", "date": "2026-08-01"},
    {"version": "v0.9.0", "minor": "0.9", "date": "2026-07-01"},
    {"version": "next", "label": "Unreleased (main)", "isNext": true}
  ]
}`),
			Mode: 0644,
		},
		"versions/v1.0.0/tutorials/hello.md":       {Data: []byte("# Hello v1\n\nWelcome to v1."), Mode: 0644},
		"versions/v1.0.0/concepts/architecture.md": {Data: []byte("# Architecture v1\n"), Mode: 0644},
		"versions/v1.0.0/tutorials/new-in-v1.md":   {Data: []byte("# New in v1\n\nAdded in v1.0.0."), Mode: 0644},
		"versions/v0.9.0/tutorials/hello.md":       {Data: []byte("# Hello v0.9\n\nWelcome to v0.9."), Mode: 0644},
		"versions/v0.9.0/concepts/architecture.md": {Data: []byte("# Architecture v0.9\n"), Mode: 0644},
		"versions/next/tutorials/hello.md":         {Data: []byte("# Hello next\n\nUnreleased."), Mode: 0644},
		"versions/next/concepts/architecture.md":   {Data: []byte("# Architecture next\n"), Mode: 0644},
		"what-is-rezuscloud.md":                    {Data: []byte("# What is RezusCloud\n\nOverview."), Mode: 0644},
	}
}

func setupVersionedDocsStore(t *testing.T) {
	t.Helper()
	store, err := docs.NewEmbeddedStore(versionedTestFS())
	require.NoError(t, err)
	DocsStore = store
}

func TestDocsVersion_ExplicitVersion(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	req := httptest.NewRequest("GET", "/docs/v1.0.0/tutorials/hello", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestDocsVersion_LatestAlias(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	req := httptest.NewRequest("GET", "/docs/latest/tutorials/hello", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestDocsVersion_NextAlias(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	req := httptest.NewRequest("GET", "/docs/next/tutorials/hello", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestDocsVersion_UnversionedBackwardCompat(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	// Unversioned path should serve latest (no redirect)
	req := httptest.NewRequest("GET", "/docs/tutorials/hello", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
	assert.NotEqual(t, 301, resp.StatusCode)
}

func TestDocsVersion_InTreeSharedDoc(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	// In-tree doc should be accessible at all version prefixes
	for _, prefix := range []string{"v1.0.0", "latest", "next", ""} {
		t.Run(prefix, func(t *testing.T) {
			url := "/docs/"
			if prefix != "" {
				url += prefix + "/"
			}
			url += "what-is-rezuscloud"
			req := httptest.NewRequest("GET", url, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, 200, resp.StatusCode)
		})
	}
}

func TestDocsVersion_DocMissingInOlderVersion(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	// new-in-v1 doesn't exist in v0.9.0 → 404 with "view latest" link
	req := httptest.NewRequest("GET", "/docs/v0.9.0/tutorials/new-in-v1", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestDocsVersion_UnknownVersion(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	// Unknown version → 404
	req := httptest.NewRequest("GET", "/docs/v9.9.9/tutorials/hello", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 404, resp.StatusCode)
}

func TestDocsVersion_VersionAwareRedirect(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	// Version-prefixed redirect
	req := httptest.NewRequest("GET", "/docs/latest/getting-started/index", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 301, resp.StatusCode)
	assert.Equal(t, "/docs/latest/tutorials/install-and-first-cluster", resp.Header.Get("Location"))
}

func TestDocsVersion_VersionAwareRedirect_ExplicitVersion(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	req := httptest.NewRequest("GET", "/docs/v1.0.0/getting-started/index", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 301, resp.StatusCode)
	assert.Equal(t, "/docs/v1.0.0/tutorials/install-and-first-cluster", resp.Header.Get("Location"))
}

func TestDocsVersion_UnversionedRedirectStillWorks(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	// Unversioned redirect (no version prefix)
	req := httptest.NewRequest("GET", "/docs/getting-started/index", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 301, resp.StatusCode)
	assert.Equal(t, "/docs/tutorials/install-and-first-cluster", resp.Header.Get("Location"))
}

func TestDocsVersion_VersionRootRedirects(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	// /docs/latest/ → redirect to first doc in latest
	req := httptest.NewRequest("GET", "/docs/latest", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should redirect (301) to a doc within the latest version
	assert.Equal(t, 301, resp.StatusCode)
	loc := resp.Header.Get("Location")
	assert.Contains(t, loc, "/docs/latest/")
}

func TestParseVersionPrefix(t *testing.T) {
	tests := []struct {
		input   string
		version string
		rest    string
	}{
		{"v0.0.1-106/tutorials/install", "v0.0.1-106", "tutorials/install"},
		{"v1.0.0/concepts/arch", "v1.0.0", "concepts/arch"},
		{"latest/tutorials/install", "latest", "tutorials/install"},
		{"next/tutorials/install", "next", "tutorials/install"},
		{"tutorials/install", "", "tutorials/install"},
		{"what-is-rezuscloud", "", "what-is-rezuscloud"},
		// Bare version aliases (no trailing path)
		{"latest", "latest", ""},
		{"next", "next", ""},
		{"v1.0.0", "v1.0.0", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, r := parseVersionPrefix(tt.input)
			assert.Equal(t, tt.version, v)
			assert.Equal(t, tt.rest, r)
		})
	}
}

func TestDocsVersion_SwitcherRenders(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	req := httptest.NewRequest("GET", "/docs/latest/tutorials/hello", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)
	assert.Contains(t, html, `id="version-select"`, "version switcher should render")
	assert.Contains(t, html, "v1.0.0", "switcher should list latest version")
	assert.Contains(t, html, "v0.9.0", "switcher should list older version")
	assert.Contains(t, html, "next (unreleased)", "switcher should label next")
}

func TestDocsVersion_CanonicalLink(t *testing.T) {
	setupVersionedDocsStore(t)
	app := setupApp()

	// Versioned page should have canonical pointing to /docs/latest/
	req := httptest.NewRequest("GET", "/docs/v1.0.0/tutorials/hello", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	html := string(body)
	assert.Contains(t, html, `rel="canonical"`, "canonical link should be present")
	assert.Contains(t, html, `https://rezus.cloud/docs/latest/tutorials/hello`,
		"canonical should point to latest version")
}

func TestDocsVersion_SwitcherNotOnLegacyStore(t *testing.T) {
	// Legacy store (no manifest) should not render the switcher
	DocsStore = nil
	fsys := fstest.MapFS{
		"tutorials/hello.md": {Data: []byte("# Hello\n"), Mode: 0644},
	}
	store, err := docs.NewEmbeddedStore(fsys)
	require.NoError(t, err)
	DocsStore = store

	app := setupApp()
	req := httptest.NewRequest("GET", "/docs/tutorials/hello", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	assert.NotContains(t, string(body), `id="version-select"`,
		"switcher should not render on legacy store")
}
