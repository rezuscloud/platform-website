package docs

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testVersionedFS creates a filesystem with two versions + shared in-tree docs.
func testVersionedFS() fstest.MapFS {
	return fstest.MapFS{
		// Manifest
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
		// v1.0.0 docs
		"versions/v1.0.0/tutorials/hello.md":       {Data: []byte("# Hello v1\n\nWelcome to v1."), Mode: 0644},
		"versions/v1.0.0/concepts/architecture.md": {Data: []byte("# Architecture v1\n"), Mode: 0644},
		"versions/v1.0.0/tutorials/new-in-v1.md":   {Data: []byte("# New in v1\n\nAdded in v1.0.0."), Mode: 0644},
		// v0.9.0 docs (older — no new-in-v1)
		"versions/v0.9.0/tutorials/hello.md":       {Data: []byte("# Hello v0.9\n\nWelcome to v0.9."), Mode: 0644},
		"versions/v0.9.0/concepts/architecture.md": {Data: []byte("# Architecture v0.9\n"), Mode: 0644},
		// next (unreleased main)
		"versions/next/tutorials/hello.md":       {Data: []byte("# Hello next\n\nUnreleased."), Mode: 0644},
		"versions/next/concepts/architecture.md": {Data: []byte("# Architecture next\n"), Mode: 0644},
		// Shared in-tree docs (platform-website's own)
		"what-is-rezuscloud.md":         {Data: []byte("# What is RezusCloud\n\nOverview."), Mode: 0644},
		"contributing/sourcing-docs.md": {Data: []byte("# Sourcing Docs\n"), Mode: 0644},
	}
}

func TestVersionedStore_HasVersions(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	assert.True(t, store.HasVersions())
	assert.Equal(t, "v1.0.0", store.LatestVersion())
}

func TestVersionedStore_ResolveVersion(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	tests := []struct {
		input string
		want  string
		err   bool
	}{
		{"latest", "v1.0.0", false},
		{"", "v1.0.0", false}, // empty resolves to latest
		{"next", "next", false},
		{"v1.0.0", "v1.0.0", false},
		{"v0.9.0", "v0.9.0", false},
		{"v9.9.9", "", true}, // unknown
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := store.ResolveVersion(tt.input)
			if tt.err {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestVersionedStore_GetVersioned_VersionSpecific(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	// v1.0.0 has "new-in-v1" that v0.9.0 doesn't
	doc, found := store.GetVersioned("v1.0.0", "tutorials/new-in-v1.md")
	require.True(t, found)
	assert.Contains(t, doc.HTML, "Added in v1.0.0")

	// v0.9.0 should NOT have it
	_, found = store.GetVersioned("v0.9.0", "tutorials/new-in-v1.md")
	assert.False(t, found)
}

func TestVersionedStore_GetVersioned_DistinctContentPerVersion(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	// Same path, different content per version
	v1doc, found := store.GetVersioned("v1.0.0", "tutorials/hello.md")
	require.True(t, found)
	assert.Contains(t, v1doc.HTML, "Welcome to v1")

	v09doc, found := store.GetVersioned("v0.9.0", "tutorials/hello.md")
	require.True(t, found)
	assert.Contains(t, v09doc.HTML, "Welcome to v0.9")

	nextDoc, found := store.GetVersioned("next", "tutorials/hello.md")
	require.True(t, found)
	assert.Contains(t, nextDoc.HTML, "Unreleased")
}

func TestVersionedStore_GetVersioned_InTreeFallback(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	// what-is-rezuscloud.md is an in-tree doc — available at all versions
	for _, version := range []string{"v1.0.0", "v0.9.0", "next"} {
		doc, found := store.GetVersioned(version, "what-is-rezuscloud.md")
		assert.True(t, found, "in-tree doc should be found at version %s", version)
		assert.Contains(t, doc.Title, "What is RezusCloud")
	}
}

func TestVersionedStore_AllDocs_VersionScoped(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	// v1.0.0 should have: tutorials/hello, concepts/architecture, tutorials/new-in-v1,
	//                    what-is-rezuscloud (shared), contributing/sourcing-docs (shared)
	v1docs := store.AllDocs("v1.0.0")
	assert.GreaterOrEqual(t, len(v1docs), 5, "v1.0.0 should have version + shared docs")

	// v0.9.0 should NOT have new-in-v1
	v09docs := store.AllDocs("v0.9.0")
	for _, d := range v09docs {
		assert.NotEqual(t, "tutorials/new-in-v1.md", d.Path, "v0.9.0 should not have new-in-v1")
	}
}

func TestVersionedStore_AvailableVersions(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	versions := store.AvailableVersions()
	assert.Len(t, versions, 3)
	assert.Equal(t, "v1.0.0", versions[0].Version)
	assert.Equal(t, "v0.9.0", versions[1].Version)
	assert.True(t, versions[2].IsNext)
}

func TestVersionedStore_Categories_VersionScoped(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	cats := store.Categories("v1.0.0")
	assert.Contains(t, cats, "tutorials")
	assert.Contains(t, cats, "concepts")
	assert.Contains(t, cats, "") // in-tree root docs have empty category

	byCat := store.DocsByCategory("v1.0.0")
	assert.NotEmpty(t, byCat["tutorials"])
	assert.NotEmpty(t, byCat["concepts"])
}

func TestVersionedStore_Get_LegacyResolvesLatest(t *testing.T) {
	store, err := NewEmbeddedStore(testVersionedFS())
	require.NoError(t, err)

	// Get() without version should resolve to latest
	doc, found := store.Get("tutorials/hello.md")
	require.True(t, found)
	assert.Contains(t, doc.HTML, "Welcome to v1")
}

// --- Legacy mode tests (no manifest) ---

func TestLegacyStore_NoManifest(t *testing.T) {
	fsys := fstest.MapFS{
		// No versions/manifest.json
		"tutorials/hello.md":    {Data: []byte("# Hello\n"), Mode: 0644},
		"what-is-rezuscloud.md": {Data: []byte("# Overview\n"), Mode: 0644},
	}

	store, err := NewEmbeddedStore(fsys)
	require.NoError(t, err)

	assert.False(t, store.HasVersions())
	assert.Equal(t, "", store.LatestVersion())

	// ResolveVersion returns "" in legacy mode
	v, err := store.ResolveVersion("latest")
	require.NoError(t, err)
	assert.Equal(t, "", v)

	// GetVersioned with "" version falls back to sharedDocs
	doc, found := store.GetVersioned("", "tutorials/hello.md")
	require.True(t, found)
	assert.Contains(t, doc.Title, "Hello")
}
