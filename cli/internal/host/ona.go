package host

// newOna returns a HostAdapter for Ona.
// Tier-3: speculative path; community-confirmable.
func newOna() HostAdapter {
	return &genericAdapter{
		id:         "ona",
		name:       "Ona",
		homepage:   "https://ona.com",
		homeSubdir: ".ona",
		caps:       Caps{Skills: true},
	}
}
