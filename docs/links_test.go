package docs

import (
	"strings"
	"testing"
	"testing/fstest"
)

func linksTestStore(t *testing.T) *Store {
	t.Helper()
	fsys := fstest.MapFS{
		"external/rezuscloud/what-is-rezuscloud.md":                  {Data: []byte("# What is RezusCloud\n")},
		"external/rezuscloud/tutorials/install-and-first-cluster.md": {Data: []byte("# Tutorial\n")},
		"external/rezuscloud/concepts/architecture.md":               {Data: []byte("# Architecture\n")},
		"external/rezuscloud/how-to/deploy-on-oci.md":                {Data: []byte("# OCI\n")},
		"external/rezuscloud/reference/cli.md":                       {Data: []byte("# CLI\n")},
	}
	s, err := NewEmbeddedStore(fsys)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func htmlWithLinks(t *testing.T, s *Store, docPath string, links ...string) string {
	t.Helper()
	var b strings.Builder
	b.WriteString("<p>body</p>\n")
	for _, href := range links {
		b.WriteString(`<a href="` + href + `">link</a>`)
	}
	return s.rewriteLinks(b.String(), docPath)
}

func TestRewriteLinks_GitHubWikiDeepLinks(t *testing.T) {
	s := linksTestStore(t)
	got := htmlWithLinks(t, s, "what-is-rezuscloud.md",
		"https://github.com/rezuscloud/rezuscloud/wiki/concepts/architecture.md",
		"https://github.com/rezuscloud/rezuscloud/wiki/getting-started/index.md",       // DocRedirects map
		"https://github.com/rezuscloud/rezuscloud/wiki/adr/0001-what-rezuscloud-is.md", // not served
		"https://github.com/rezuscloud/rezuscloud/wiki",                                // bare, no .md
	)

	if !strings.Contains(got, `href="/docs/concepts/architecture"`) {
		t.Errorf("served wiki target not rewritten: %s", got)
	}
	if !strings.Contains(got, `href="/docs/tutorials/install-and-first-cluster"`) {
		t.Errorf("redirect-mapped target not rewritten: %s", got)
	}
	if !strings.Contains(got, `href="https://github.com/rezuscloud/rezuscloud/wiki/adr/0001-what-rezuscloud-is.md"`) {
		t.Errorf("unserved target must be left as-is: %s", got)
	}
	if !strings.Contains(got, `href="https://github.com/rezuscloud/rezuscloud/wiki"`) {
		t.Errorf("bare wiki link must be left as-is: %s", got)
	}
}

func TestRewriteLinks_RelativeMarkdownTargets(t *testing.T) {
	s := linksTestStore(t)

	// From a root-level page, repo-relative targets resolve from the root.
	root := htmlWithLinks(t, s, "what-is-rezuscloud.md",
		"tutorials/install-and-first-cluster.md")
	if !strings.Contains(root, `href="/docs/tutorials/install-and-first-cluster"`) {
		t.Errorf("repo-relative target not rewritten: %s", root)
	}

	// From a nested page, page-relative targets climb with "..".
	nested := htmlWithLinks(t, s, "how-to/deploy-on-oci.md",
		"../concepts/architecture.md",
		"../reference/missing.md", // not served
	)
	if !strings.Contains(nested, `href="/docs/concepts/architecture"`) {
		t.Errorf("page-relative target not rewritten: %s", nested)
	}
	if !strings.Contains(nested, `href="../reference/missing.md"`) {
		t.Errorf("unserved relative target must be left as-is: %s", nested)
	}

	// A relative target escaping the docs root is left alone.
	escape := htmlWithLinks(t, s, "how-to/deploy-on-oci.md", "../../../outside.md")
	if !strings.Contains(escape, `href="../../../outside.md"`) {
		t.Errorf("escaping target must be left as-is: %s", escape)
	}
}

func TestRewriteLinks_NonDocTargetsUntouched(t *testing.T) {
	s := linksTestStore(t)
	got := htmlWithLinks(t, s, "what-is-rezuscloud.md",
		"https://github.com/rezuscloud/rezuscloud/blob/main/docs/adr/README.md", // ADR blob
		"https://demo.rezus.cloud",
		"#section-anchor",
		"mailto:admin@rezus.cloud",
		"/docs/already-clean",
	)

	for _, want := range []string{
		`href="https://github.com/rezuscloud/rezuscloud/blob/main/docs/adr/README.md"`,
		`href="https://demo.rezus.cloud"`,
		`href="#section-anchor"`,
		`href="mailto:admin@rezus.cloud"`,
		`href="/docs/already-clean"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected untouched %s in: %s", want, got)
		}
	}
}

func TestBuildDoc_HealsLinksEndToEnd(t *testing.T) {
	fsys := fstest.MapFS{
		"external/rezuscloud/what-is-rezuscloud.md": {Data: []byte(
			"# What is RezusCloud\n\nGo to [the tutorial](https://github.com/rezuscloud/rezuscloud/wiki/tutorials/install-and-first-cluster.md).\n")},
		"external/rezuscloud/tutorials/install-and-first-cluster.md": {Data: []byte("# Tutorial\n")},
	}
	s, err := NewEmbeddedStore(fsys)
	if err != nil {
		t.Fatal(err)
	}
	doc, ok := s.Get("what-is-rezuscloud.md")
	if !ok {
		t.Fatal("doc not indexed")
	}
	if !strings.Contains(doc.HTML, `href="/docs/tutorials/install-and-first-cluster"`) {
		t.Fatalf("indexed HTML not healed: %s", doc.HTML)
	}
	if strings.Contains(doc.HTML, "github.com") {
		t.Fatalf("github link survived in indexed HTML: %s", doc.HTML)
	}
}

func TestBuildDoc_HealsLinksAcrossWikisRegardlessOfWalkOrder(t *testing.T) {
	// Regression: lexical walk order indexes the linking wiki before the
	// linked one (alpha < rezuscloud). Healing must not depend on it.
	fsys := fstest.MapFS{
		"external/alpha-wiki/what-is-x.md": {Data: []byte(
			"# What is X\n\nSee [architecture](https://github.com/rezuscloud/rezuscloud/wiki/concepts/architecture.md).\n")},
		"external/rezuscloud/concepts/architecture.md": {Data: []byte("# Architecture\n")},
	}
	s, err := NewEmbeddedStore(fsys)
	if err != nil {
		t.Fatal(err)
	}
	doc, ok := s.Get("what-is-x.md")
	if !ok {
		t.Fatal("doc not indexed")
	}
	if !strings.Contains(doc.HTML, `href="/docs/concepts/architecture"`) {
		t.Fatalf("cross-wiki link not healed (walk-order regression): %s", doc.HTML)
	}
}
