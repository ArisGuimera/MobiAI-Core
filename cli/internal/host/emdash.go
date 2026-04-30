package host

// newEmdash returns a HostAdapter for Emdash.
// Tier-3: speculative path; community-confirmable.
func newEmdash() HostAdapter {
	return &genericAdapter{
		id:         "emdash",
		name:       "Emdash",
		homepage:   "https://emdash.dev",
		homeSubdir: ".emdash",
		caps:       Caps{Skills: true},
	}
}
