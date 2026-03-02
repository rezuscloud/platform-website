package handlers

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rezuscloud/platform-website/version"
)

func TestAPIVersion(t *testing.T) {
	// Set test version
	originalVersion := version.Version
	originalGitCommit := version.GitCommit
	originalBuildTime := version.BuildTime

	version.Version = "1.2.3"
	version.GitCommit = "abc123"
	version.BuildTime = "2026-03-02"

	defer func() {
		version.Version = originalVersion
		version.GitCommit = originalGitCommit
		version.BuildTime = originalBuildTime
	}()

	app := fiber.New()
	app.Get("/api/version", APIVersion)

	req := httptest.NewRequest("GET", "/api/version", nil)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var info version.Info
	err = json.Unmarshal(body, &info)
	require.NoError(t, err)

	assert.Equal(t, "1.2.3", info.Version)
	assert.Equal(t, "abc123", info.GitCommit)
	assert.Equal(t, "2026-03-02", info.BuildTime)
}
