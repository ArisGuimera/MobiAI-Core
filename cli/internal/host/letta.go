package host

// newLetta returns a HostAdapter for Letta.
// Tier-3: speculative path; community-confirmable.
func newLetta() HostAdapter {
	return &genericAdapter{
		id:         "letta",
		name:       "Letta",
		homepage:   "https://www.letta.com",
		homeSubdir: ".letta",
		caps:       Caps{Skills: true},
	}
}
