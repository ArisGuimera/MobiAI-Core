package host

// newRooCode returns a HostAdapter for Roo Code (VS Code extension).
// Tier-2: best-effort detection at ~/.roo-code/.
func newRooCode() HostAdapter {
	return &genericAdapter{
		id:         "roo-code",
		name:       "Roo Code",
		homepage:   "https://github.com/RooCodeInc/Roo-Code",
		homeSubdir: ".roo-code",
		caps:       Caps{Skills: true},
	}
}
