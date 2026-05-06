package host

// newGoose returns a HostAdapter for Block Goose (open-source agent).
// Tier-2: best-effort detection at ~/.config/goose/.
func newGoose() HostAdapter {
	return &genericAdapter{
		id:         "goose",
		name:       "Goose",
		homepage:   "https://block.github.io/goose",
		homeSubdir: ".config/goose",
		caps:       Caps{Skills: true},
	}
}
