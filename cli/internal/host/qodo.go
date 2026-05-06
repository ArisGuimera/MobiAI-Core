package host

// newQodo returns a HostAdapter for Qodo.
// Tier-3: speculative path; community-confirmable.
func newQodo() HostAdapter {
	return &genericAdapter{
		id:         "qodo",
		name:       "Qodo",
		homepage:   "https://www.qodo.ai",
		homeSubdir: ".qodo",
		caps:       Caps{Skills: true},
	}
}
