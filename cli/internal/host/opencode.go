package host

// newOpenCode returns a HostAdapter for OpenCode.
// Tier-2: best-effort detection at ~/.agents/.
func newOpenCode() HostAdapter {
	return &genericAdapter{
		id:         "opencode",
		name:       "OpenCode",
		homepage:   "https://opencode.ai",
		homeSubdir: ".agents",
		caps:       Caps{Skills: true},
	}
}
