package host

// newCommandCode returns a HostAdapter for Command Code.
// Tier-3: speculative path; community-confirmable.
func newCommandCode() HostAdapter {
	return &genericAdapter{
		id:         "command-code",
		name:       "Command Code",
		homepage:   "https://commandcode.dev",
		homeSubdir: ".command-code",
		caps:       Caps{Skills: true},
	}
}
