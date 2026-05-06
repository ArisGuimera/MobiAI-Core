package host

// newPiebald returns a HostAdapter for Piebald.
// Tier-3: speculative path; community-confirmable.
func newPiebald() HostAdapter {
	return &genericAdapter{
		id:         "piebald",
		name:       "Piebald",
		homepage:   "https://piebald.ai",
		homeSubdir: ".piebald",
		caps:       Caps{Skills: true},
	}
}
