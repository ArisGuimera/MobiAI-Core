package host

// newCodex returns a HostAdapter for OpenAI Codex. Skills only — no
// host-specific hook/command surface.
func newCodex() HostAdapter {
	return &genericAdapter{
		id:         "codex",
		name:       "Codex",
		homepage:   "https://openai.com/codex",
		homeSubdir: ".codex",
		caps:       Caps{Skills: true},
	}
}
