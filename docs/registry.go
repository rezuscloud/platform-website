package docs

// RepoConfig defines a source repository whose docs/ folder is indexed.
type RepoConfig struct {
	// Name is the GitHub repository name under rezuscloud.
	Name string `json:"name"`

	// DisplayName is the human-readable project name shown in navigation.
	DisplayName string `json:"displayName"`

	// Description is a one-line summary for the repo index.
	Description string `json:"description"`

	// DocsPath is the subdirectory within the repo containing docs.
	// Defaults to "docs" if empty.
	DocsPath string `json:"docsPath,omitempty"`

	// VersionTag is the git tag used for GitHub "View source" links.
	// If empty, defaults to "main".
	VersionTag string `json:"versionTag,omitempty"`
}

// Registry lists all documentation sources.
var Registry = []RepoConfig{
	{
		Name:        "platform-website",
		DisplayName: "Platform Website",
		Description: "Marketing website for RezusCloud Enterprise Kubernetes Platform",
		DocsPath:    "docs",
	},
	{
		Name:        "rezusctl",
		DisplayName: "RezusCloud CLI",
		Description: "Single tool for managing the full lifecycle of a RezusCloud Personal Cloud",
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
