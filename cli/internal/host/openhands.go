package host

// newOpenHands returns a HostAdapter for OpenHands.
// Tier-3: speculative path; community-confirmable.
func newOpenHands() HostAdapter {
	return &genericAdapter{
		id:         "openhands",
		name:       "OpenHands",
		homepage:   "https://github.com/All-Hands-AI/OpenHands",
		homeSubdir: ".openhands",
		caps:       Caps{Skills: true},
	}
}
