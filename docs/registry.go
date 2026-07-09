package docs

// RepoConfig defines a source repository for GitHub links.
// The actual documentation content lives in the website's /docs folder,
// organized by topic (tutorials, concepts, reference, adr).
type RepoConfig struct {
	// Name is the GitHub repository name under rezuscloud.
	Name string

	// DisplayName is the human-readable project name.
	DisplayName string

	// DocsPath is the subdirectory within the repo containing source docs.
	// Used for constructing GitHub edit/view URLs.
	DocsPath string

	// VersionTag is the git branch used for GitHub links.
	// If empty, defaults to "main".
	VersionTag string
}

// Registry lists source repositories for attribution and GitHub links.
// The sidebar shows categories (from directory structure), not repo names.
//
// Source-of-truth policy: product documentation lives in the rezuscloud
// repository and the llm-wiki. The platform-website repo carries only a
// minimal set of high-level concept pages and ADRs that define the
// website's own voice. Everything else is fetched at build time.
var Registry = []RepoConfig{
	{
		Name:        "platform-website",
		DisplayName: "Platform Website",
		DocsPath:    "docs",
		VersionTag:  "master",
	},
	{
		Name:        "rezuscloud",
		DisplayName: "RezusCloud",
		DocsPath:    "docs",
	},
}

// GitHubBaseURL returns the base GitHub URL for viewing docs.
func (r RepoConfig) GitHubBaseURL() string {
	branch := r.VersionTag
	if branch == "" {
		branch = "main"
	}
	return "https://github.com/rezuscloud/" + r.Name + "/blob/" + branch + "/" + r.DocsPath
}

// GitHubEditURL returns the base GitHub URL for editing docs.
func (r RepoConfig) GitHubEditURL() string {
	branch := r.VersionTag
	if branch == "" {
		branch = "main"
	}
	return "https://github.com/rezuscloud/" + r.Name + "/edit/" + branch + "/" + r.DocsPath
}
