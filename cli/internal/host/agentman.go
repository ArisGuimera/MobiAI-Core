package host

// newAgentman returns a HostAdapter for Agentman.
// Tier-3: speculative path; community-confirmable.
func newAgentman() HostAdapter {
	return &genericAdapter{
		id:         "agentman",
		name:       "Agentman",
		homepage:   "https://agentman.ai",
		homeSubdir: ".agentman",
		caps:       Caps{Skills: true},
	}
}
