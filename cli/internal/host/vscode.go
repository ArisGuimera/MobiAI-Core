package host

// newVSCode returns a HostAdapter for VS Code (with Copilot or other AI extensions).
// Tier-2: best-effort detection at ~/.vscode/.
func newVSCode() HostAdapter {
	return &genericAdapter{
		id:         "vscode",
		name:       "VS Code",
		homepage:   "https://code.visualstudio.com",
		homeSubdir: ".vscode",
		caps:       Caps{Skills: true},
	}
}
