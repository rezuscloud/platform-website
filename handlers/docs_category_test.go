package handlers

import (
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/rezuscloud/platform-website/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupDocsStore injects a deterministic docs store (wikis are gitignored, so
// tests must not depend on a prior fetch-docs.sh run) and restores the global
// on cleanup.
func setupDocsStore(t *testing.T) {
	t.Helper()
	fsys := fstest.MapFS{
		"external/rezuscloud/tutorials/install-and-first-cluster.md": {Data: []byte("# Tutorial\n")},
		"external/rezuscloud/how-to/add-bare-metal-node.md":          {Data: []byte("# Bare metal\n")},
		"external/rezuscloud/how-to/deploy-on-oci.md":                {Data: []byte("# OCI\n")},
		"external/rezuscloud/reference/cli.md":                       {Data: []byte("# CLI\n")},
		// adr/ is not an allowed category — must never surface as a root.
		"external/rezuscloud/adr/0001-what-rezuscloud-is.md": {Data: []byte("# ADR\n")},
	}
	store, err := docs.NewEmbeddedStore(fsys)
	require.NoError(t, err)

	orig := DocsStore
	DocsStore = store
	t.Cleanup(func() { DocsStore = orig })
}

func TestDocsCategoryRootRedirectsToFirstDoc(t *testing.T) {
	setupDocsStore(t)
	app := setupApp()

	t.Run("category root 302s to the section's first doc in sidebar order", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/docs/how-to", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 302, resp.StatusCode,
			"category roots redirect, they are not 404")
		assert.Equal(t, "/docs/how-to/add-bare-metal-node", resp.Header.Get("Location"),
			"target is the first doc of the category by path sort")
	})

	t.Run("trailing slash redirects too", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/docs/reference/", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 302, resp.StatusCode)
		assert.Equal(t, "/docs/reference/cli", resp.Header.Get("Location"))
	})

	t.Run("unknown path still 404s", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/docs/how-to2", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode,
			"only real category names redirect")
	})

	t.Run("non-allowed category dir is not a root", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/docs/adr", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 404, resp.StatusCode,
			"the Diátaxis allowlist gates category roots too")
	})

	t.Run("real doc still serves", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/docs/how-to/deploy-on-oci", nil)
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, 200, resp.StatusCode)
	})
}
