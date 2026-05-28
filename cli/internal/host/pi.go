package host

// newPi returns a HostAdapter for pi.
// Tier-3: speculative path; community-confirmable.
func newPi() HostAdapter {
	return &genericAdapter{
		id:         "pi",
		name:       "pi",
		homepage:   "https://pi.ai",
		homeSubdir: ".agents",
		caps:       Caps{Skills: true},
	}
}
