package host

// newGemini returns a HostAdapter for Google Gemini CLI. Gemini CLI sticks
// to plain skills via the agentskills.io standard — no Claude-specific
// hooks/commands.
func newGemini() HostAdapter {
	return &genericAdapter{
		id:         "gemini",
		name:       "Gemini CLI",
		homepage:   "https://github.com/google-gemini/gemini-cli",
		homeSubdir: ".gemini",
		caps:       Caps{Skills: true, MCPs: true},
	}
}
