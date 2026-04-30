package host

// newClaudeCode returns a HostAdapter for Anthropic Claude Code.
// Claude Code stores skills under ~/.claude/skills/ and additionally
// supports hooks and slash commands; those bonus capabilities are flagged
// here but actual hook installation is deferred to a later phase.
func newClaudeCode() HostAdapter {
	return &genericAdapter{
		id:         "claude-code",
		name:       "Claude Code",
		homepage:   "https://claude.com/code",
		homeSubdir: ".claude",
		caps:       Caps{Skills: true, Hooks: true, Commands: true, MCPs: true},
	}
}
