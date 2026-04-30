package host

// newMux returns a HostAdapter for Mux.
// Tier-3: speculative path; community-confirmable.
func newMux() HostAdapter {
	return &genericAdapter{
		id:         "mux",
		name:       "Mux",
		homepage:   "https://mux.com",
		homeSubdir: ".mux",
		caps:       Caps{Skills: true},
	}
}
