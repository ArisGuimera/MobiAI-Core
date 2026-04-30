package host

// newFirebender returns a HostAdapter for Firebender.
// Tier-3: speculative path; community-confirmable.
func newFirebender() HostAdapter {
	return &genericAdapter{
		id:         "firebender",
		name:       "Firebender",
		homepage:   "https://firebender.com",
		homeSubdir: ".firebender",
		caps:       Caps{Skills: true},
	}
}
