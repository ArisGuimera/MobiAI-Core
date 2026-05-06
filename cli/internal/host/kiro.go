package host

// newKiro returns a HostAdapter for Kiro.
// Tier-3: speculative path; community-confirmable.
func newKiro() HostAdapter {
	return &genericAdapter{
		id:         "kiro",
		name:       "Kiro",
		homepage:   "https://kiro.dev",
		homeSubdir: ".kiro",
		caps:       Caps{Skills: true},
	}
}
