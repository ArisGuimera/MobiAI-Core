package host

// newVTCode returns a HostAdapter for VT Code.
// Tier-3: speculative path; community-confirmable.
func newVTCode() HostAdapter {
	return &genericAdapter{
		id:         "vt-code",
		name:       "VT Code",
		homepage:   "https://github.com/vinhnx/vtcode",
		homeSubdir: ".vt-code",
		caps:       Caps{Skills: true},
	}
}
