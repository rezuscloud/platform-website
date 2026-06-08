// Generated helper functions for live.templ
// NOTE: This file is NOT auto-generated. Edit manually.

package sections

import "strings"

// shortHost returns a shortened version of a hostname for display.
func shortHost(name string) string {
	if len(name) > 20 {
		return name[:8] + ".." + name[len(name)-8:]
	}
	return name
}

// contains checks if substr is in s.
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
