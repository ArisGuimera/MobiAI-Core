package host

// newCursor returns a HostAdapter for Cursor. Cursor mirrors Claude Code's
// extension surface (skills + hooks + commands + MCPs).
func newCursor() HostAdapter {
	return &genericAdapter{
		id:         "cursor",
		name:       "Cursor",
		homepage:   "https://cursor.com",
		homeSubdir: ".cursor",
		caps:       Caps{Skills: true, Hooks: true, Commands: true, MCPs: true},
	}
}
