package host

// newFactory returns a HostAdapter for Factory.
// Tier-3: speculative path; community-confirmable.
func newFactory() HostAdapter {
	return &genericAdapter{
		id:         "factory",
		name:       "Factory",
		homepage:   "https://factory.ai",
		homeSubdir: ".factory",
		caps:       Caps{Skills: true},
	}
}
