package host

// newTRAE returns a HostAdapter for TRAE.
// Tier-3: speculative path; community-confirmable.
func newTRAE() HostAdapter {
	return &genericAdapter{
		id:         "trae",
		name:       "TRAE",
		homepage:   "https://trae.ai",
		homeSubdir: ".trae",
		caps:       Caps{Skills: true},
	}
}
