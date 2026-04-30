package host

import (
	"os"
	"path/filepath"
)

// genericAdapter implements every HostAdapter method against a configurable
// home subdirectory. Tier-1 adapters are thin factories around it.
type genericAdapter struct {
	id         string
	name       string
	homepage   string
	homeSubdir string // e.g., ".claude"
	caps       Caps
}

func (a *genericAdapter) ID() string         { return a.id }
func (a *genericAdapter) Name() string       { return a.name }
func (a *genericAdapter) Homepage() string   { return a.homepage }
func (a *genericAdapter) Capabilities() Caps { return a.caps }

func (a *genericAdapter) hostDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, a.homeSubdir)
}

func (a *genericAdapter) SkillsDir() string {
	return filepath.Join(a.hostDir(), "skills")
}

func (a *genericAdapter) Detect() DetectResult {
	dir := a.hostDir()
	if dir == "" {
		return DetectResult{Found: false, Searched: []string{"<unresolved home>"}}
	}
	info, err := os.Stat(dir)
	if err == nil && info.IsDir() {
		return DetectResult{Found: true, Path: dir, Searched: []string{dir}}
	}
	return DetectResult{Found: false, Searched: []string{dir}}
}
