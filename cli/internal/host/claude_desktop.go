package host

// newClaudeDesktop returns a HostAdapter for Claude Desktop.
// Tier-3: speculative path; community-confirmable.
func newClaudeDesktop() HostAdapter {
	return &genericAdapter{
		id:         "claude-desktop",
		name:       "Claude Desktop",
		homepage:   "https://claude.ai/download",
		homeSubdir: ".claude-desktop",
		caps:       Caps{Skills: true},
	}
}
