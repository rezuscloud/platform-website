package pages

import "strings"

// docPathToURL converts a doc path like "adr/0001-foo.md" to a URL path "/docs/repo/adr/0001-foo".
func docPathToURL(repoName, docPath string) string {
	return "/docs/" + repoName + "/" + strings.TrimSuffix(docPath, ".md")
}
