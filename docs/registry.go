package docs

// RepoConfig defines a documentation source.
type RepoConfig struct {
	// Name is the repository name under rezuscloud (e.g. "rezuscloud"). For a
	// wiki source, this is the base repo name; the actual source is its
	// <name>.wiki.git repository.
	Name string

	// DisplayName is the human-readable project name.
	DisplayName string

	// IsWiki reports whether the source is the repo's GitHub wiki
	// (<name>.wiki.git) rather than a docs/ directory in the repo itself.
	IsWiki bool
}

// Registry lists documentation sources.
//
// Source-of-truth policy: user-facing documentation is authored in the project
// wikis and fetched at build time (see scripts/fetch-docs.sh). The repositories
// carry only ADRs and repo-level context (CONTEXT.md, PRODUCT.md, …), which are
// never served here. Because the wikis contain only user docs — no ADRs, no
// archived decision history — internal content cannot leak onto the public docs
// site regardless of what a repo adds later (the store's Diátaxis allowlist is
// the second line of defense).
var Registry = []RepoConfig{
	{Name: "rezuscloud", DisplayName: "RezusCloud", IsWiki: true},
	{Name: "platform-website", DisplayName: "Platform Website", IsWiki: true},
}

// GitHubBaseURL returns the base URL for viewing a source.
func (r RepoConfig) GitHubBaseURL() string {
	if r.IsWiki {
		return "https://github.com/rezuscloud/" + r.Name + "/wiki"
	}
	return "https://github.com/rezuscloud/" + r.Name + "/blob/main/docs"
}

// GitHubEditURL returns the URL for editing a source. Wikis are edited via
// their web UI; we link to the wiki home rather than a per-file URL (GitHub
// wiki page URLs do not reliably preserve subdirectory structure).
func (r RepoConfig) GitHubEditURL() string {
	if r.IsWiki {
		return "https://github.com/rezuscloud/" + r.Name + "/wiki"
	}
	return "https://github.com/rezuscloud/" + r.Name + "/edit/main/docs"
}
