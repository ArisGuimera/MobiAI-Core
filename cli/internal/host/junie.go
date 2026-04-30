package host

// newJunie returns a HostAdapter for JetBrains Junie.
// Tier-2: best-effort detection at ~/.junie/.
func newJunie() HostAdapter {
	return &genericAdapter{
		id:         "junie",
		name:       "Junie",
		homepage:   "https://www.jetbrains.com/junie",
		homeSubdir: ".junie",
		caps:       Caps{Skills: true},
	}
}
