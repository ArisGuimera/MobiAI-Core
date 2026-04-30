package host

// newFastAgent returns a HostAdapter for fast-agent.
// Tier-3: speculative path; community-confirmable.
func newFastAgent() HostAdapter {
	return &genericAdapter{
		id:         "fast-agent",
		name:       "fast-agent",
		homepage:   "https://github.com/evalstate/fast-agent",
		homeSubdir: ".fast-agent",
		caps:       Caps{Skills: true},
	}
}
