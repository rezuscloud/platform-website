package pages

import "strings"

// trimExt removes the .md extension from a path.
func trimExt(path string) string {
	return strings.TrimSuffix(path, ".md")
}
