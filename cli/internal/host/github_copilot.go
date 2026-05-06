package host

// newGitHubCopilot returns a HostAdapter for the GitHub Copilot CLI.
// Tier-2: best-effort detection at ~/.copilot/.
func newGitHubCopilot() HostAdapter {
	return &genericAdapter{
		id:         "github-copilot",
		name:       "GitHub Copilot",
		homepage:   "https://github.com/features/copilot",
		homeSubdir: ".copilot",
		caps:       Caps{Skills: true},
	}
}
