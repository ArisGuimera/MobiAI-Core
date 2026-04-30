package host

// newMistralVibe returns a HostAdapter for Mistral AI Vibe.
// Tier-3: speculative path; community-confirmable.
func newMistralVibe() HostAdapter {
	return &genericAdapter{
		id:         "mistral-vibe",
		name:       "Mistral AI Vibe",
		homepage:   "https://mistral.ai",
		homeSubdir: ".mistral-vibe",
		caps:       Caps{Skills: true},
	}
}
