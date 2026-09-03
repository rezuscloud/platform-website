package docs

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestStripFirstH1(t *testing.T) {
	t.Run("strips the first H1 only", func(t *testing.T) {
		in := "<h1 id=\"title\">Title</h1>\n<p>intro</p>\n<h1>Another</h1>\n"
		got := stripFirstH1(in)
		if strings.Contains(got, "<h1 id=\"title\">") {
			t.Fatalf("first H1 not stripped: %q", got)
		}
		if !strings.Contains(got, "<h1>Another</h1>") {
			t.Fatalf("later H1 should be preserved: %q", got)
		}
		if !strings.Contains(got, "<p>intro</p>") {
			t.Fatalf("body content lost: %q", got)
		}
	})

	t.Run("no H1 is a no-op", func(t *testing.T) {
		in := "<p>intro</p>\n<h2>Section</h2>\n"
		if got := stripFirstH1(in); got != in {
			t.Fatalf("want unchanged, got %q", got)
		}
	})

	t.Run("unclosed H1 is a no-op", func(t *testing.T) {
		in := "<h1 id=\"x\">Broken"
		if got := stripFirstH1(in); got != in {
			t.Fatalf("want unchanged, got %q", got)
		}
	})
}

func TestBuildDoc_DropsDuplicateTitleHeading(t *testing.T) {
	// Simulates a wiki page whose markdown starts with an H1: the template
	// renders doc.Title as the page <h1>, so the article body must not repeat it.
	src := []byte("# Documentation Standards\n\nStatus: adopted.\n\n## 1. Sections\n")
	doc := buildDoc("documentation-standards.md", src, "rezuscloud")

	if doc.Title != "Documentation Standards" {
		t.Fatalf("title = %q, want %q", doc.Title, "Documentation Standards")
	}
	if strings.Contains(doc.HTML, "<h1") {
		t.Fatalf("rendered HTML still contains an H1 (duplicate title): %q", doc.HTML)
	}
	if !strings.Contains(doc.HTML, "Status: adopted.") {
		t.Fatalf("body content lost: %q", doc.HTML)
	}
}

func TestBuildDoc_FallbackTitleKeepsNoH1(t *testing.T) {
	// A page without an H1 falls back to a filename-derived title and its
	// HTML is untouched (nothing to strip).
	src := []byte("Just some prose.\n")
	doc := buildDoc("some-page.md", src, "")

	if doc.Title != "some page" {
		t.Fatalf("title = %q, want %q", doc.Title, "some page")
	}
	if !strings.Contains(doc.HTML, "Just some prose.") {
		t.Fatalf("body content lost: %q", doc.HTML)
	}
}

func TestFirstDocInCategory(t *testing.T) {
	fsys := fstest.MapFS{
		"external/rezuscloud/how-to/deploy-on-oci.md":                {Data: []byte("# B\n")},
		"external/rezuscloud/how-to/add-bare-metal-node.md":          {Data: []byte("# A\n")},
		"external/rezuscloud/tutorials/install-and-first-cluster.md": {Data: []byte("# C\n")},
		// Not an allowed category — never indexed, so never a root.
		"external/rezuscloud/adr/0001-what-rezuscloud-is.md": {Data: []byte("# D\n")},
	}
	s, err := NewEmbeddedStore(fsys)
	if err != nil {
		t.Fatal(err)
	}

	if got, ok := s.FirstDocInCategory("how-to"); !ok || got != "how-to/add-bare-metal-node.md" {
		t.Fatalf("FirstDocInCategory(how-to) = %q, %v; want first by path sort, true", got, ok)
	}
	if _, ok := s.FirstDocInCategory("adr"); ok {
		t.Fatal("non-allowed category must have no docs, ok=false")
	}
	if _, ok := s.FirstDocInCategory(""); ok {
		t.Fatal("empty category must be false (root handled by DocsIndex)")
	}
	if _, ok := s.FirstDocInCategory("nonexistent"); ok {
		t.Fatal("unknown category must be false")
	}
}
