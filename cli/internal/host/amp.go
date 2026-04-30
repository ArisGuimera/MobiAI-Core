package host

// newAmp returns a HostAdapter for Amp.
// Tier-3: speculative path; community-confirmable.
func newAmp() HostAdapter {
	return &genericAdapter{
		id:         "amp",
		name:       "Amp",
		homepage:   "https://ampcode.com",
		homeSubdir: ".amp",
		caps:       Caps{Skills: true},
	}
}
