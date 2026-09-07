package docs

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testStore(t *testing.T, files map[string]string) *Store {
	t.Helper()
	mapfs := fstest.MapFS{}
	for path, data := range files {
		mapfs[path] = &fstest.MapFile{Data: []byte(data)}
	}
	s, err := NewEmbeddedStore(mapfs)
	require.NoError(t, err)
	return s
}

const introPage = "# What is Rezus.cloud\n\nThe product intro.\n"
const standardsPage = "# Documentation standards\n\nMeta page.\n"

func TestLandingPath_PrefersProductIntro(t *testing.T) {
	s := testStore(t, map[string]string{
		"external/platform-website/documentation-standards.md": standardsPage,
		"external/platform-website/what-is-rezuscloud.md":      introPage,
	})

	assert.Equal(t, "what-is-rezuscloud", s.LandingPath())
}

func TestLandingPath_FallsBackToFirstIndexedPage(t *testing.T) {
	s := testStore(t, map[string]string{
		"external/platform-website/documentation-standards.md": standardsPage,
		"external/platform-website/other.md":                   "# Other\n",
	})

	// No landing candidate present: first indexed page wins.
	assert.Equal(t, "documentation-standards", s.LandingPath())
}

func TestLandingPath_EmptyStore(t *testing.T) {
	s := testStore(t, map[string]string{})

	assert.Equal(t, "", s.LandingPath())
}
