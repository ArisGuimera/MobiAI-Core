package host

// newAutohand returns a HostAdapter for Autohand.
// Tier-3: speculative path; community-confirmable.
func newAutohand() HostAdapter {
	return &genericAdapter{
		id:         "autohand",
		name:       "Autohand",
		homepage:   "https://autohand.ai",
		homeSubdir: ".autohand",
		caps:       Caps{Skills: true},
	}
}
