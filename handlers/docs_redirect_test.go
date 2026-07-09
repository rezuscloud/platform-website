package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsRedirect_RenamedPaths(t *testing.T) {
	app := setupApp()

	tests := []struct {
		oldPath     string
		wantStatus  int
		wantLocPath string
	}{
		// Diátaxis reorg: getting-started/ → tutorials/ + concepts/
		{"/docs/getting-started/index", 301, "/docs/tutorials/install-and-first-cluster"},
		{"/docs/getting-started/multi-cluster", 301, "/docs/concepts/multi-cluster"},
		{"/docs/getting-started/what-is-rezuscloud", 301, "/docs/what-is-rezuscloud"},
		{"/docs/getting-started/install", 301, "/docs/tutorials/install-and-first-cluster"},
		// integrations/ → how-to/
		{"/docs/integrations/home-assistant", 301, "/docs/how-to/integrate-home-assistant"},
	}

	for _, tt := range tests {
		t.Run(tt.oldPath, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.oldPath, nil)
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode,
				"%s: expected %d, got %d", tt.oldPath, tt.wantStatus, resp.StatusCode)
			assert.Equal(t, tt.wantLocPath, resp.Header.Get("Location"),
				"%s: expected redirect to %s, got %s", tt.oldPath, tt.wantLocPath, resp.Header.Get("Location"))
		})
	}
}

func TestDocsRedirect_NonRenamedPathNoRedirect(t *testing.T) {
	app := setupApp()

	// A path that doesn't exist and isn't in the redirect map should 404, not redirect.
	req := httptest.NewRequest("GET", "/docs/nonexistent/deep/path", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, 301, resp.StatusCode, "non-renamed paths must not redirect")
	assert.NotEqual(t, 302, resp.StatusCode, "non-renamed paths must not redirect")
}

func TestDocsRedirect_ValidDocNotRedirected(t *testing.T) {
	app := setupApp()

	// A valid existing doc should render (200 or 404 if store is empty in tests),
	// but must NOT redirect.
	req := httptest.NewRequest("GET", "/docs/tutorials/install-and-first-cluster", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.NotEqual(t, 301, resp.StatusCode, "valid docs must not redirect")
	assert.NotEqual(t, 302, resp.StatusCode, "valid docs must not redirect")
}
