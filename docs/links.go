package docs

import (
	"path"
	"regexp"
	"strings"
)

// Wiki pages are authored against the GitHub wiki: deep links like
// https://github.com/rezuscloud/rezuscloud/wiki/concepts/architecture.md and
// repo-relative targets like ../concepts/architecture.md. Neither reads well
// for visitors: GitHub serves .md wiki URLs as raw markdown (or an empty view
// when the page doesn't exist there), and relative targets resolve against the
// current page URL, so they 404 unless the depth happens to line up.
//
// rewriteLinks heals this at render time: when a link target resolves to a doc
// this site serves (honoring DocRedirects), the href becomes the clean on-site
// URL. Everything else — ADR blob links (intentionally not served here),
// external sites, anchors, mailto — is left untouched. An unresolvable .md
// target is also left as-is: the renderer never invents destinations.

// githubWikiLinkRe matches deep links into the rezuscloud wikis, capturing the
// repo-relative page path (e.g. "concepts/architecture.md").
var githubWikiLinkRe = regexp.MustCompile(`^https://github\.com/rezuscloud/[^/]+/wiki/(.+\.md)$`)

// hrefRe matches href attributes in the HTML this package renders (goldmark
// emits double-quoted, entity-escaped attributes; wiki targets are plain
// paths/URLs and need no unescaping).
var hrefRe = regexp.MustCompile(`href="([^"]+)"`)

// resolveLinkTarget maps a link target to a served doc path (without ".md"),
// or reports false when the site cannot serve that target.
func (s *Store) resolveLinkTarget(target, docPath string) (string, bool) {
	var candidate string
	switch {
	case strings.HasPrefix(target, "#"):
		return "", false
	case githubWikiLinkRe.MatchString(target):
		candidate = githubWikiLinkRe.FindStringSubmatch(target)[1]
	case strings.HasSuffix(target, ".md") && !strings.Contains(target, "://"):
		// Repo-relative or page-relative target: resolve against the doc's
		// directory (store paths are .md-suffixed, e.g. how-to/deploy-on-oci.md).
		candidate = path.Clean(path.Join(path.Dir(docPath), target))
		if candidate == "." || strings.HasPrefix(candidate, "..") {
			return "", false
		}
	default:
		return "", false
	}

	candidate = strings.TrimSuffix(candidate, ".md")
	if mapped := Redirect(candidate); mapped != "" {
		candidate = mapped
	}
	if _, ok := s.Get(candidate + ".md"); ok {
		return candidate, true
	}
	return "", false
}

// rewriteLinks rewrites every href whose target resolves to a served doc.
func (s *Store) rewriteLinks(html, docPath string) string {
	return hrefRe.ReplaceAllStringFunc(html, func(attr string) string {
		href := hrefRe.FindStringSubmatch(attr)[1]
		if target, ok := s.resolveLinkTarget(href, docPath); ok {
			return `href="/docs/` + target + `"`
		}
		return attr
	})
}
