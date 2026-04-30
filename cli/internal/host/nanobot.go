package host

// newNanobot returns a HostAdapter for nanobot.
// Tier-3: speculative path; community-confirmable.
func newNanobot() HostAdapter {
	return &genericAdapter{
		id:         "nanobot",
		name:       "nanobot",
		homepage:   "https://github.com/nanobot-ai/nanobot",
		homeSubdir: ".nanobot",
		caps:       Caps{Skills: true},
	}
}
