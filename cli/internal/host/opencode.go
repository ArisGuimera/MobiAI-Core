package host

// newOpenCode returns a HostAdapter for OpenCode.
// Tier-2: best-effort detection at ~/.opencode/.
func newOpenCode() HostAdapter {
	return &genericAdapter{
		id:         "opencode",
		name:       "OpenCode",
		homepage:   "https://opencode.ai",
		homeSubdir: ".opencode",
		caps:       Caps{Skills: true},
	}
}
