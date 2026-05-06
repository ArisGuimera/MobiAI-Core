package host

// newCodex returns a HostAdapter for the OpenAI Codex CLI (the open-source
// terminal coding agent at github.com/openai/codex, released 2025).
// Not to be confused with the older OpenAI Codex code-completion model
// (deprecated 2023). Skills only — no host-specific hook/command surface.
func newCodex() HostAdapter {
	return &genericAdapter{
		id:         "codex",
		name:       "Codex",
		homepage:   "https://github.com/openai/codex",
		homeSubdir: ".codex",
		caps:       Caps{Skills: true},
	}
}
